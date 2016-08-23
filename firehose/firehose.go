package firehose

import (
	"crypto/tls"
	"time"

	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

func CreateFirehoseChan(DopplerEndpoint string, Token string, subId string, skipSSLValidation bool, keepAlive time.Duration) <-chan *events.Envelope {
	consumer.KeepAlive = keepAlive
	connection := consumer.New(DopplerEndpoint, &tls.Config{InsecureSkipVerify: skipSSLValidation}, nil)
	connection.SetDebugPrinter(ConsoleDebugPrinter{})
	msgChan, errorChan := connection.Firehose(subId, Token)
	go func() {
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
