package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/events"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v1"
	"log"
	"os"
	"time"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Bool()
	domain            = kingpin.Flag("domain", "Domain of your CF installation.").Default("10.244.0.34.xip.io").String()
	syslogServer      = kingpin.Flag("syslog-server", "Syslog server.").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Bool()
	wantedEvents      = kingpin.Flag("events", fmt.Sprintf("Comma seperated list of events you would like. Valid options are %s", events.GetListAuthorizedEventEvents())).Default("LogMessage").String()
	boldDatabasePath  = kingpin.Flag("boltdb-path", "Bolt Database path ").Default("my.db").String()
	tickerTime        = kingpin.Flag("cc-pull-time", "CloudController Pooling time in sec").Default("60s").Duration()
)

func main() {
	kingpin.Version("0.0.2 - ba541ca")
	kingpin.Parse()

	apiEndpoint := fmt.Sprintf("https://api.%s", *domain)
	uaaEndpoint := fmt.Sprintf("https://uaa.%s", *domain)
	dopplerEndpoint := fmt.Sprintf("wss://doppler.%s", *domain)

	c := cfclient.Config{
		ApiAddress:        apiEndpoint,
		LoginAddress:      uaaEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient := cfclient.NewClient(&c)

	//Use bolt for in-memory  - file caching
	db, err := bolt.Open(*boldDatabasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}
	defer db.Close()
	caching.SetCfClient(cfClient)
	caching.SetAppDb(db)
	caching.CreateBucket()

	// Ticker Pooling the CC every X sec
	ccPooling := time.NewTicker(*tickerTime)

	go func() {
		for range ccPooling.C {
			caching.GetAllApp()

		}

	}()

	selectedEvents := events.GetSelectedEvents(*wantedEvents)

	logging.SetupLogging(*syslogServer, *debug)

	//Let's Update the database the first time

	log.Println("Staring filling app/space/org cache.")
	caching.GetAllApp()
	log.Println("Done filling cache, I will now start processing events!")

	token := cfClient.GetToken()

	firehose := firehose.CreateFirehoseChan(dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)

	events.RouteEvents(firehose, selectedEvents)
}
