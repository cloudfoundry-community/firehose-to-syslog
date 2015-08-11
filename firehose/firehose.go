package firehose

import (
	"crypto/tls"
	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
)

func CreateFirehoseChan(DopplerEndpoint string, Token string, subId string, skipSSLValidation bool) chan *events.Envelope {
	connection := noaa.NewConsumer(DopplerEndpoint, &tls.Config{InsecureSkipVerify: skipSSLValidation}, nil)

	connection.SetDebugPrinter(ConsoleDebugPrinter{})

	msgChan := make(chan *events.Envelope)
	go func() {
		errorChan := make(chan error)
		defer close(msgChan)

		defer func() {
			if r := recover(); r != nil {
				log.LogError("Recovered in CreateFirehoseChan Thread!", r)
			}
		}()

		go connection.Firehose(subId, Token, msgChan, errorChan)

		for err := range errorChan {
			log.LogError("Firehose Error!", err.Error())
		}
	}()
	return msgChan
}

type ConsoleDebugPrinter struct{}

func (c ConsoleDebugPrinter) Print(title, dump string) {
	log.LogStd(title, false)
	log.LogStd(dump, false)
}
