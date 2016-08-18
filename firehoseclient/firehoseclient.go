package firehoseclient

import (
	"crypto/tls"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
	"time"
)

type FirehoseNozzle struct {
	errs         <-chan error
	messages     <-chan *events.Envelope
	consumer     *consumer.Consumer
	eventRouting *eventRouting.EventRouting
	config       *FirehoseConfig
	cfClient     *cfclient.Client
}

type FirehoseConfig struct {
	TrafficControllerURL   string
	InsecureSSLSkipVerify  bool
	IdleTimeoutSeconds     time.Duration
	FirehoseSubscriptionID string
}

type CfClientTokenRefresh struct {
	cfClient *cfclient.Client
}

func NewFirehoseNozzle(cfClient *cfclient.Client, eventRouting *eventRouting.EventRouting, firehoseconfig *FirehoseConfig) *FirehoseNozzle {
	return &FirehoseNozzle{
		errs:         make(<-chan error),
		messages:     make(<-chan *events.Envelope),
		eventRouting: eventRouting,
		config:       firehoseconfig,
		cfClient:     cfClient,
	}
}

func (f *FirehoseNozzle) Start() error {
	f.consumeFirehose()
	err := f.routeEvent()
	return err
}

func (f *FirehoseNozzle) consumeFirehose() {
	f.consumer = consumer.New(
		f.config.TrafficControllerURL,
		&tls.Config{InsecureSkipVerify: f.config.InsecureSSLSkipVerify},
		nil)
	refresher := CfClientTokenRefresh{cfClient: f.cfClient}
	f.consumer.RefreshTokenFrom(&refresher)
	f.consumer.SetIdleTimeout(time.Duration(f.config.IdleTimeoutSeconds) * time.Second)
	f.messages, f.errs = f.consumer.Firehose(f.config.FirehoseSubscriptionID, "")
}

func (f *FirehoseNozzle) routeEvent() error {
	for {
		select {
		case envelope := <-f.messages:
			f.eventRouting.RouteEvent(envelope)
		case err := <-f.errs:
			f.handleError(err)
			return err
		}
	}
}

func (f *FirehoseNozzle) handleError(err error) {

	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		logging.LogError("Normal Websocket Closure: %v", err)
	case websocket.IsCloseError(err, websocket.ClosePolicyViolation):
		logging.LogError("Error while reading from the firehose: %v", err)
		logging.LogError("Disconnected because nozzle couldn't keep up. Please try scaling up the nozzle.", nil)

	default:
		logging.LogError("Error while reading from the firehose: %v", err)
	}

	logging.LogError("Closing connection with traffic controller due to %v", err)
	f.consumer.Close()
}

func (f *FirehoseNozzle) handleMessage(envelope *events.Envelope) {
	if envelope.GetEventType() == events.Envelope_CounterEvent && envelope.CounterEvent.GetName() == "TruncatingBuffer.DroppedMessages" && envelope.GetOrigin() == "doppler" {
		logging.LogStd("We've intercepted an upstream message which indicates that the nozzle or the TrafficController is not keeping up. Please try scaling up the nozzle.", true)
	}
}

func (ct *CfClientTokenRefresh) RefreshAuthToken() (string, error) {
	return ct.cfClient.GetToken(), nil
}
