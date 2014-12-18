package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/SpringerPE/firehose-to-syslog/token"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log/syslog"
	"os"
)

var (
	debug            = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Bool()
	uaaEndpoint      = kingpin.Flag("uaa-endpoint", "UAA endpoint.").Required().String()
	dopplerEndpoint  = kingpin.Flag("doppler-endpoint", "UAA endpoint.").Required().String()
	syslogServer     = kingpin.Flag("syslog-server", "Syslog server.").String()
	subscriptionId   = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").String()
	firehoseUser     = kingpin.Flag("firehose-user", "User with firehose permissions.").Default("doppler").String()
	firehosePassword = kingpin.Flag("firehose-password", "Password for firehose user.").Default("doppler").String()
)

func CreateFirehoseChan(DopplerEndpoint string, Token string, subId string) chan *events.Envelope {
	connection := noaa.NewConsumer(DopplerEndpoint, nil, nil)
	msgChan := make(chan *events.Envelope)
	go func() {
		errorChan := make(chan error)
		defer close(msgChan)
		defer close(errorChan)

		go connection.Firehose(subId, Token, msgChan, errorChan, nil)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()
	return msgChan
}

func FilterEvents(in chan *events.Envelope) chan *events.Envelope {
	out := make(chan *events.Envelope)
	go func() {
		defer close(out)
		for msg := range in {
			switch msg.GetEventType().String() {
			case "LogMessage":
				out <- msg
			}
		}
	}()
	return out
}

func Logger(in chan *events.Envelope) {
	for msg := range in {
		logmsg := msg.GetLogMessage()
		app_id := logmsg.GetAppId()

		log.WithFields(log.Fields{
			"cf_app_id":       app_id,
			"source_type":     logmsg.GetSourceType(),
			"message_type":    logmsg.GetMessageType().String(),
			"source_instance": logmsg.GetSourceInstance(),
		}).Info(string(logmsg.GetMessage()))
	}
}

func setupLogging(syslogServer string, debug bool) {
	log.SetFormatter(&log.JSONFormatter{})
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	if syslogServer != "" {
		hook, err := logrus_syslog.NewSyslogHook("tcp", syslogServer, syslog.LOG_INFO, "doppler")
		if err != nil {
			log.Error("Unable to connect to syslog server.")
		} else {
			log.AddHook(hook)
		}
	}
}

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()

	setupLogging(*syslogServer, *debug)
	token := token.GetToken(*uaaEndpoint, *firehoseUser, *firehosePassword)

	firehose := CreateFirehoseChan(*dopplerEndpoint, token, *subscriptionId)
	logs := FilterEvents(firehose)
	Logger(logs)
}
