package firehose

import (
	"crypto/tls"
	"fmt"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
	"os"
)

func CreateFirehoseChan(DopplerEndpoint string, Token string, subId string, skipSSLValidation bool) chan *events.Envelope {
	connection := noaa.NewConsumer(DopplerEndpoint, &tls.Config{InsecureSkipVerify: skipSSLValidation}, nil)
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
