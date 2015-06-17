package firehose

import (
	"crypto/tls"
	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
)

// CreateFirehoseChan creates a firehose channel that we consume events from.
func CreateFirehoseChan(DopplerEndpoint string, Token string, subID string, skipSSLValidation bool) chan *events.Envelope {
	connection := noaa.NewConsumer(DopplerEndpoint, &tls.Config{InsecureSkipVerify: skipSSLValidation}, nil)

	connection.SetDebugPrinter(consoleDebugPrinter{})

	msgChan := make(chan *events.Envelope)
	go func() {
		errorChan := make(chan error)
		defer close(msgChan)

		defer func() {
			if r := recover(); r != nil {
				log.LogError("Recovered in CreateFirehoseChan Thread!", r)
			}
		}()

		go connection.Firehose(subID, Token, msgChan, errorChan, nil)

		for err := range errorChan {
			log.LogError("Firehose Error!", err.Error())
		}
	}()
	return msgChan
}

type consoleDebugPrinter struct{}

// Print is a function for a noaa connection
func (c consoleDebugPrinter) Print(title, dump string) {
	log.LogStd(title, false)
	log.LogStd(dump, false)
}
