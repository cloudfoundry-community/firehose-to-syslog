package firehoseclient

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/stats"

	gendiodes "code.cloudfoundry.org/diodes"
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
	messages       <-chan *events.Envelope
	consumer       *consumer.Consumer
	eventRouting   eventRouting.EventRouting
	config         *FirehoseConfig
	uaaRefresher   consumer.TokenRefresher
	envelopeBuffer *diodes.OneToOneEnvelope
	doneCh         chan struct{}
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

func NewFirehoseNozzle(uaaR *uaatokenrefresher.UAATokenRefresher,
	eventRouting eventRouting.EventRouting,
	firehoseconfig *FirehoseConfig,
	stats *stats.Stats) *FirehoseNozzle {
	return &FirehoseNozzle{
		errs:         make(<-chan error),
		messages:     make(<-chan *events.Envelope),
		eventRouting: eventRouting,
		config:       firehoseconfig,
		uaaRefresher: uaaR,
		envelopeBuffer: diodes.NewOneToOneEnvelope(firehoseconfig.BufferSize, gendiodes.AlertFunc(func(missed int) {
			logging.LogError("Missed Logs ", missed)
		})),
		doneCh: make(chan struct{}),
		Stats:  stats,
	}
}

func (f *FirehoseNozzle) Stop() {
	logging.LogStd("Stopping Channel ", true)
	close(f.doneCh)
}

func (f *FirehoseNozzle) Start() error {
	defer f.Stop()
	f.consumeFirehose()
	go f.ReadLogsBuffer()
	err := f.routeEvent()
	return err
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

func (f *FirehoseNozzle) ReadLogsBuffer() {
	for {
		select {
		case <-f.doneCh:
			return
		default:
			envelope := f.envelopeBuffer.Next()
			f.handleMessage(envelope)
			f.eventRouting.RouteEvent(envelope)
			f.Stats.Dec(stats.SubInputBuffer)
		}
	}
}

func (f *FirehoseNozzle) routeEvent() error {
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
				return err
			}
		case <-f.doneCh:
			return fmt.Errorf("Closing Go routine")
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
