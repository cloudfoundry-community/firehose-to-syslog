package main

import (
	"fmt"
	"github.com/cloudfoundry-community/firehose-to-syslog/events"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Bool()
	domain            = kingpin.Flag("domain", "Domain of your CF installation.").Required().String()
	syslogServer      = kingpin.Flag("syslog-server", "Syslog server.").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Bool()
	wantedEvents      = kingpin.Flag("events", fmt.Sprintf("Comma seperated list of events you would like. Valid options are %s", events.GetListAuthorizedEventEvents())).Default("LogMessage").String()
)

func main() {
	kingpin.Version("0.0.2 - ba541ca")
	kingpin.Parse()

	apiEndpoint := fmt.Sprintf("https://api.%s", *domain)
	uaaEndpoint := fmt.Sprintf("https://uaa.%s", *domain)
	dopplerEndpoint := fmt.Sprintf("wss://doppler.%s", *domain)

	c := cfclient.Config{
		ApiAddress:   apiEndpoint,
		LoginAddress: uaaEndpoint,
		Username:     *user,
		Password:     *password,
	}
	cfClient := cfclient.NewClient(&c)

	selectedEvents := events.GetSelectedEvents(*wantedEvents)

	logging.SetupLogging(*syslogServer, *debug)

	token := cfClient.GetToken()
	firehose := firehose.CreateFirehoseChan(dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)

	events.RouteEvents(firehose, selectedEvents)
}
