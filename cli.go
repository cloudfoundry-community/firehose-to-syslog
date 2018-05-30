package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehoseclient"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/firehose-to-syslog/stats"
	"github.com/cloudfoundry-community/firehose-to-syslog/uaatokenrefresher"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Default("false").Envar("DEBUG").Bool()
	apiEndpoint       = kingpin.Flag("api-endpoint", "Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io").Envar("API_ENDPOINT").Required().String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").Envar("DOPPLER_ENDPOINT").String()
	syslogServer      = kingpin.Flag("syslog-server", "Syslog server.").Envar("SYSLOG_ENDPOINT").String()
	syslogProtocol    = kingpin.Flag("syslog-protocol", "Syslog protocol (tcp/udp/tcp+tls).").Default("tcp").Envar("SYSLOG_PROTOCOL").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").Envar("FIREHOSE_SUBSCRIPTION_ID").String()
	clientID          = kingpin.Flag("client-id", "Client ID.").Envar("FIREHOSE_CLIENT_ID").Required().String()
	clientSecret      = kingpin.Flag("client-secret", "Client secret.").Envar("FIREHOSE_CLIENT_SECRET").Required().String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").Envar("SKIP_SSL_VALIDATION").Bool()
	keepAlive         = kingpin.Flag("fh-keep-alive", "Keep Alive duration for the firehose consumer").Default("25s").Envar("FH_KEEP_ALIVE").Duration()
	minRetryDelay     = kingpin.Flag("min-retry-delay", "Doppler Cloud Foundry Doppler min. retry delay duration").Default("500ms").Envar("MIN_RETRY_DELAY").Duration()
	maxRetryDelay     = kingpin.Flag("max-retry-delay", "Doppler Cloud Foundry Doppler max. retry delay duration").Default("1m").Envar("MAX_RETRY_DELAY").Duration()
	maxRetryCount     = kingpin.Flag("max-retry-count", "Doppler Cloud Foundry Doppler max. retry Count duration").Default("1000").Envar("MAX_RETRY_COUNT").Int()
	bufferSize        = kingpin.Flag("logs-buffer-size", "Number of envelope to be buffered").Default("10000").Envar("LOGS_BUFFER_SIZE").Int()
	wantedEvents      = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).Default("LogMessage").Envar("EVENTS").String()
	statServer        = kingpin.Flag("enable-stats-server", "Will enable stats server on 8080").Default("true").Envar("ENABLE_STATS_SERVER").Bool()
	boltDatabasePath  = kingpin.Flag("boltdb-path", "Bolt Database path ").Default("my.db").Envar("BOLTDB_PATH").String()
	tickerTime        = kingpin.Flag("cc-pull-time", "CloudController Polling time in sec").Default("60s").Envar("CF_PULL_TIME").Duration()
	requestLimit      = kingpin.Flag("cc-rps", "CloudController Polling request by second").Default("50").Envar("CF_RPS").Int()
	extraFields       = kingpin.Flag("extra-fields", "Extra fields you want to annotate your events with, example: '--extra-fields=env:dev,something:other ").Default("").Envar("EXTRA_FIELDS").String()
	cfOrgs            = kingpin.Flag("cf-orgs", "stream only certain cloudfoundry orgs' LogMessage'--cf-orgs=env:dev,something:org1|org2").Default("").Envar("CF_ORGS").String()
	modeProf          = kingpin.Flag("mode-prof", "Enable profiling mode, one of [cpu, mem, block]").Default("").Envar("MODE_PROF").String()
	pathProf          = kingpin.Flag("path-prof", "Set the Path to write profiling file").Default("").Envar("PATH_PROF").String()
	logFormatterType  = kingpin.Flag("log-formatter-type", "Log formatter type to use. Valid options are text, json. If none provided, defaults to json.").Envar("LOG_FORMATTER_TYPE").String()
	certPath          = kingpin.Flag("cert-pem-syslog", "Certificate Pem file").Envar("CERT_PEM").Default("").String()
	ignoreMissingApps = kingpin.Flag("ignore-missing-apps", "Enable throttling on cache lookup for missing apps").Envar("IGNORE_MISSING_APPS").Default("false").Bool()
)

const (
	ExitCodeOk    = 0
	ExitCodeError = 1 + iota
)

var (
	version = "0.0.0"
)

// CLI is the command line object
type CLI struct {
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	kingpin.Version(version)
	kingpin.Parse()

	//Setup Logging
	loggingClient := logging.NewLogging(*syslogServer, *syslogProtocol, *logFormatterType, *certPath, *debug)
	logging.LogStd(fmt.Sprintf("Starting firehose-to-syslog %s ", version), true)
	//
	// if *modeProf != "" {
	// 	switch *modeProf {
	// 	case "cpu":
	// 		defer profile.Start(profile.CPUProfile, profile.ProfilePath(*pathProf)).Stop()
	// 	case "mem":
	// 		defer profile.Start(profile.MemProfile, profile.ProfilePath(*pathProf)).Stop()
	// 	case "block":
	// 		defer profile.Start(profile.BlockProfile, profile.ProfilePath(*pathProf)).Stop()
	// 	default:
	// 		// do nothing
	// 	}
	// }

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		ClientID:          *clientID,
		ClientSecret:      *clientSecret,
		SkipSslValidation: *skipSSLValidation,
		UserAgent:         "firehose-to-syslog/" + version,
	}
	cfClient, err := cfclient.NewClient(&c)
	if err != nil {
		logging.LogError("New Client: ", err)
		return ExitCodeError

	}
	if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}
	fmt.Println(cfClient.Endpoint.DopplerEndpoint)
	logging.LogStd(fmt.Sprintf("Using %s as doppler endpoint", cfClient.Endpoint.DopplerEndpoint), true)

	//Creating Caching
	var cachingClient caching.Caching
	if caching.IsNeeded(*wantedEvents) {
		config := &caching.CachingBoltConfig{
			Path:               *boltDatabasePath,
			IgnoreMissingApps:  *ignoreMissingApps,
			CacheInvalidateTTL: *tickerTime,
			RequestBySec:       *requestLimit,
		}
		cachingClient, err = caching.NewCachingBolt(cfClient, config)

		if err != nil {
			logging.LogError("Failed to create boltdb cache", err)
			return ExitCodeError
		}
	} else {
		cachingClient = caching.NewCachingEmpty()
	}

	if err := cachingClient.Open(); err != nil {
		logging.LogError("Error open cache: ", err)
		return ExitCodeError
	}
	defer cachingClient.Close()

	//Adding Stats
	statistic := stats.NewStats()
	go statistic.PerSec()

	////Starting Http Server for Stats
	if *statServer {
		Server := &stats.Server{
			Logger: log.New(os.Stdout, "", 1),
			Stats:  statistic,
		}

		go Server.Start()
	}

	//Creating Events
	eventFilters := []eventRouting.EventFilter{eventRouting.HasIgnoreField, eventRouting.NotInCertainOrgs(*cfOrgs)}
	events := eventRouting.NewEventRouting(cachingClient, loggingClient, statistic, eventFilters)
	err = events.SetupEventRouting(*wantedEvents)
	if err != nil {
		logging.LogError("Error setting up event routing: ", err)
		return ExitCodeError
	}

	//Set extrafields if needed
	events.SetExtraFields(*extraFields)

	uaaRefresher, err := uaatokenrefresher.NewUAATokenRefresher(
		cfClient.Endpoint.AuthEndpoint,
		*clientID,
		*clientSecret,
		*skipSSLValidation,
	)

	if err != nil {
		logging.LogError(fmt.Sprint("Failed connecting to Get token from UAA..", err), "")
	}

	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   cfClient.Endpoint.DopplerEndpoint,
		InsecureSSLSkipVerify:  *skipSSLValidation,
		IdleTimeoutSeconds:     *keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
		MinRetryDelay:          *minRetryDelay,
		MaxRetryDelay:          *maxRetryDelay,
		MaxRetryCount:          *maxRetryCount,
		BufferSize:             *bufferSize,
	}

	if loggingClient.Connect() || *debug {
		logging.LogStd("Connected to Syslog Server! Connecting to Firehose...", true)
	} else {
		logging.LogError("Failed connecting to the Syslog Server...Please check settings and try again!", "")
		return ExitCodeError
	}

	firehoseClient := firehoseclient.NewFirehoseNozzle(uaaRefresher, events, firehoseConfig, statistic)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	firehoseClient.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nSignal Received, Stop reading and starting Draining...")
			firehoseClient.StopReading()
			cctx, tcancel := context.WithTimeout(context.TODO(), 30*time.Second)
			tcancel()
			firehoseClient.Draining(cctx)
			cleanupDone <- true
		}
	}()

	<-cleanupDone

	return ExitCodeOk
}
