package firehoseclient

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/stats"

	gendiodes "code.cloudfoundry.org/go-diodes"
	"github.com/cloudfoundry-community/firehose-to-syslog/diodes"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/firehose-to-syslog/uaatokenrefresher"
	"github.com/cloudfoundry/noaa/consumer"
	noaerrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"

	"github.com/gorilla/websocket"
)

type FirehoseNozzle struct {
	errs           <-chan error
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
}

type FirehoseConfig struct {
	MinRetryDelay          time.Duration
	MaxRetryDelay          time.Duration
	MaxRetryCount          int
	TrafficControllerURL   string
	InsecureSSLSkipVerify  bool
	IdleTimeoutSeconds     time.Duration
	FirehoseSubscriptionID string
	BufferSize             int
}

var (
	wg sync.WaitGroup
)

func NewFirehoseNozzle(uaaR *uaatokenrefresher.UAATokenRefresher,
	eventRouting eventRouting.EventRouting,
	firehoseconfig *FirehoseConfig,
	stats *stats.Stats) *FirehoseNozzle {
	return &FirehoseNozzle{
		errs:         make(<-chan error),
		Readerrs:     make(chan error),
		messages:     make(<-chan *events.Envelope),
		eventRouting: eventRouting,
		config:       firehoseconfig,
		uaaRefresher: uaaR,
		envelopeBuffer: diodes.NewOneToOneEnvelope(firehoseconfig.BufferSize, gendiodes.AlertFunc(func(missed int) {
			logging.LogError("Missed Logs ", missed)
		})),
		stopReading: make(chan struct{}),
		stopRouting: make(chan struct{}),
		Stats:       stats,
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
	f.consumer = consumer.New(
		f.config.TrafficControllerURL,
		&tls.Config{InsecureSkipVerify: f.config.InsecureSSLSkipVerify},
		nil)
	f.consumer.RefreshTokenFrom(f.uaaRefresher)
	f.consumer.SetIdleTimeout(f.config.IdleTimeoutSeconds)
	f.consumer.SetMinRetryDelay(f.config.MinRetryDelay)
	f.consumer.SetMaxRetryDelay(f.config.MaxRetryDelay)
	f.consumer.SetMaxRetryCount(f.config.MaxRetryCount)
	f.messages, f.errs = f.consumer.Firehose(f.config.FirehoseSubscriptionID, "")
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
			f.handleMessage(envelope)
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
			f.handleMessage(envelope)
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
		case err := <-f.errs:
			f.handleError(err)
			retryrerr := f.handleError(err)
			if !retryrerr {
				logging.LogError("RouteEvent Loop Error ", err)
				return
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

func (f *FirehoseNozzle) handleError(err error) bool {
	logging.LogError("Error while reading from the Firehose: ", err)

	switch err.(type) {
	case noaerrors.RetryError:
		switch noaRetryError := err.(noaerrors.RetryError).Err.(type) {
		case *websocket.CloseError:
			switch noaRetryError.Code {
			case websocket.ClosePolicyViolation:
				logging.LogError("Nozzle couldn't keep up. Please try scaling up the Nozzle.", err)
			}
		}
		return true
	}

	logging.LogStd("Closing connection with Firehose...", true)
	f.consumer.Close()
	return false
}

func (f *FirehoseNozzle) handleMessage(envelope *events.Envelope) {
	if envelope.GetEventType() == events.Envelope_CounterEvent && envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" && envelope.GetOrigin() == "doppler" {
		logging.LogStd("We've intercepted an upstream message which indicates that the nozzle or the TrafficController is not keeping up. Please try scaling up the nozzle.", true)
	}
}
