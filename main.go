package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/eljuanchosf/firehose-to-syslog/caching"
	"github.com/eljuanchosf/firehose-to-syslog/events"
	"github.com/eljuanchosf/firehose-to-syslog/extrafields"
	"github.com/eljuanchosf/firehose-to-syslog/firehose"
	"github.com/eljuanchosf/firehose-to-syslog/logging"
	"github.com/eljuanchosf/firehose-to-syslog/filters"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/profile"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"time"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Default("false").Bool()
	apiEndpoint       = kingpin.Flag("api-endpoint", "Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io").Required().String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").String()
	syslogServer      = kingpin.Flag("syslog-server", "Syslog server.").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").Bool()
	wantedEvents      = kingpin.Flag("events", fmt.Sprintf("Comma seperated list of events you would like. Valid options are %s", events.GetListAuthorizedEventEvents())).Default("LogMessage").String()
	boltDatabasePath  = kingpin.Flag("boltdb-path", "Bolt Database path ").Default("my.db").String()
	tickerTime        = kingpin.Flag("cc-pull-time", "CloudController Pooling time in sec").Default("60s").Duration()
	extraFields       = kingpin.Flag("extra-fields", "Extra fields you want to annotate your events with, example: --extra-fields=env:dev,something:other ").Default("").String()
	modeProf          = kingpin.Flag("mode-prof", "Enable profiling mode, one of [cpu, mem, block]").Default("").String()
	pathProf          = kingpin.Flag("path-prof", "Set the Path to write profiling file").Default("").String()
	customFilters     = kingpin.Flag("filters", "Pipe seperated whitelist filtering for messages. Possible keys: org_name, org_id, space_name, space_id, app_name, app_id. Values are comma seperated. Example: --filters=\"org_name:org1,org2|space_id:asff-12ffa,1122-dbfa-aaaa|app_name:app1\"").Default("").String()
)

const (
	version = "1.1.0 - c9af859"
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()
	logging.LogStd(fmt.Sprintf("Starting firehose-to-syslog %s ", version), true)

	logging.SetupLogging(*syslogServer, *debug)

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient := cfclient.NewClient(&c)

	if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}
	logging.LogStd(fmt.Sprintf("Using %s as doppler endpoint", cfClient.Endpoint.DopplerEndpoint), true)

	logging.LogStd("Setting up event routing!", true)
	err := events.SetupEventRouting(*wantedEvents)
	if err != nil {
		log.Fatal("Error setting up event routing: ", err)
		os.Exit(1)

	}

	//Use bolt for in-memory  - file caching
	db, err := bolt.Open(*boltDatabasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}
	defer db.Close()

	if *modeProf != "" {
		switch *modeProf {
		case "cpu":
			defer profile.Start(profile.CPUProfile, profile.ProfilePath(*pathProf)).Stop()
		case "mem":
			defer profile.Start(profile.MemProfile, profile.ProfilePath(*pathProf)).Stop()
		case "block":
			defer profile.Start(profile.BlockProfile, profile.ProfilePath(*pathProf)).Stop()
		default:
			// do nothing
		}
	}

	caching.SetCfClient(cfClient)
	caching.SetAppDb(db)
	caching.CreateBucket()

	//Let's Update the database the first time
	logging.LogStd("Start filling app/space/org cache.", true)
	apps := caching.GetAllApp()
	logging.LogStd(fmt.Sprintf("Done filling cache! Found [%d] Apps", len(apps)), true)

	// Ticker Pooling the CC every X sec
	ccPooling := time.NewTicker(*tickerTime)

	go func() {
		for range ccPooling.C {
			apps = caching.GetAllApp()
		}
	}()

	// Parse extra fields from cmd call
	extraFields, err := extrafields.ParseExtraFields(*extraFields)
	if err != nil {
		log.Fatal("Error parsing extra fields: ", err)
		os.Exit(1)
	}

	// Parse filters
	filtersToApply, err := filters.ParseFilters(*customFilters);
	if err != nil {
	  log.Fatal("Error parsing filters: ", err)
	  os.Exit(1)
	}

	if logging.Connect() || *debug {

		logging.LogStd("Connected to Syslog Server! Connecting to Firehose...", true)

		firehose := firehose.CreateFirehoseChan(cfClient.Endpoint.DopplerEndpoint, cfClient.GetToken(), *subscriptionId, *skipSSLValidation)
		if firehose != nil {
			logging.LogStd("Firehose Subscription Succesfull! Routing events...", true)
			events.RouteEvents(firehose, extraFields, filtersToApply)
		} else {
			logging.LogError("Failed connecting to Firehose...Please check settings and try again!", "")
		}

	} else {
		logging.LogError("Failed connecting to the Syslog Server...Please check settings and try again!", "")
	}
}
