package firehoseclient

import (
	"code.cloudfoundry.org/go-loggregator"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/stats"

	gendiodes "code.cloudfoundry.org/go-diodes"
	"github.com/cloudfoundry-community/firehose-to-syslog/diodes"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

type FirehoseNozzle struct {
	Readerrs       chan error
	messages       <-chan *events.Envelope
	consumer       *consumer.Consumer
	eventRouting   eventRouting.EventRouting
	config         *FirehoseConfig
	uaaRefresher   consumer.TokenRefresher
	envelopeBuffer *diodes.OneToOneEnvelope
	stopReading    chan struct{}
	stopRouting    chan struct{}
	Stats          *stats.Stats
	httpClient     doer
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type FirehoseConfig struct {
	RLPAddr                string
	InsecureSSLSkipVerify  bool
	FirehoseSubscriptionID string
	BufferSize             int
}

var (
	wg sync.WaitGroup
)

func NewFirehoseNozzle(
	eventRouting eventRouting.EventRouting,
	firehoseconfig *FirehoseConfig,
	stats *stats.Stats,
	httpClient doer,
	) *FirehoseNozzle {
	return &FirehoseNozzle{
		Readerrs:     make(chan error),
		messages:     make(<-chan *events.Envelope),
		eventRouting: eventRouting,
		config:       firehoseconfig,
		envelopeBuffer: diodes.NewOneToOneEnvelope(firehoseconfig.BufferSize, gendiodes.AlertFunc(func(missed int) {
			logging.LogError("Missed Logs ", missed)
		})),
		stopReading: make(chan struct{}),
		stopRouting: make(chan struct{}),
		Stats:       stats,
		httpClient: httpClient,
	}
}

//Start consumer and reading ingest loop
func (f *FirehoseNozzle) Start(ctx context.Context) {
	f.consumeFirehose()
	wg.Add(2)
	go f.routeEvent(ctx)
	go f.ReadLogsBuffer(ctx)
}

//Stop reading loop
func (f *FirehoseNozzle) StopReading() {
	close(f.stopRouting)
	close(f.stopReading)
	//Need to be sure both of the GoRoutine are stop
	wg.Wait()
}

func (f *FirehoseNozzle) consumeFirehose() {
	rlpGatewayClient := loggregator.NewRLPGatewayClient(
		f.config.RLPAddr,
		loggregator.WithRLPGatewayHTTPClient(f.httpClient),
		)
	a := NewV2Adapter(rlpGatewayClient)
	f.messages = a.Firehose(f.config.FirehoseSubscriptionID)
}

func (f *FirehoseNozzle) ReadLogsBuffer(ctx context.Context) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logging.LogStd("Cancel ReadLogsBuffer Goroutine", true)
			return
		case <-f.stopRouting:
			logging.LogStd("Stopping Routing Loop", true)
			return
		default:
			envelope, empty := f.envelopeBuffer.TryNext()
			if envelope == nil && !empty {
				// Brief sleep to Avoid hammering on CPU
				time.Sleep(1 * time.Millisecond)
				continue
			}
			f.eventRouting.RouteEvent(envelope)
			f.Stats.Dec(stats.SubInputBuffer)
		}
	}
}

func (f *FirehoseNozzle) Draining(ctx context.Context) {
	logging.LogStd("Starting Draining", true)
	for {
		select {
		case <-ctx.Done():
			logging.LogStd("Stopping ReadLogsBuffer Goroutine", true)
			return
		default:
			envelope, empty := f.envelopeBuffer.TryNext()
			if envelope == nil && !empty {
				logging.LogStd("Finishing Draining", true)
				return
			}
			f.eventRouting.RouteEvent(envelope)
			f.Stats.Dec(stats.SubInputBuffer)
		}
	}
}

func (f *FirehoseNozzle) routeEvent(ctx context.Context) {
	defer wg.Done()
	eventsSelected := f.eventRouting.GetSelectedEvents()
	for {
		select {
		case envelope := <-f.messages:
			//Only take what we need
			if eventsSelected[envelope.GetEventType().String()] {
				f.envelopeBuffer.Set(envelope)
				f.Stats.Inc(stats.SubInputBuffer)
			}
		case <-ctx.Done():
			logging.LogStd("Closing routing event routine", true)
			return
		case <-f.stopReading:
			logging.LogStd("Stopping Reading from Firehose", true)
			return
		}
	}
}
