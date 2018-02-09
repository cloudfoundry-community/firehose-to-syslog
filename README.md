# Disclaimer

Since **2.5.0** we stop supporting username and password for authentification.

Please use ClientId and ClientSecret.

# Firehose-to-syslog

This nifty util aggregates all the events from the firehose feature in
CloudFoundry.

	./firehose-to-syslog \
              --api-endpoint="https://api.10.244.0.34.xip.io" \
              --skip-ssl-validation \
              --debug
	....
	....
	{"cf_app_id":"c5cb762b-b7bb-44b6-97d1-2b612d4baba9","cf_app_name":"lattice","cf_org_id":"fb5777e6-e234-4832-8844-773114b505b0","cf_org_name":"GWENN","cf_origin":"firehose","cf_space_id":"3c910823-22e7-41ff-98de-094759594398","cf_space_name":"GWENN-SPACE","event_type":"LogMessage","level":"info","message_type":"OUT","msg":"Lattice-app. Says Hello. on index: 0","origin":"rep","source_instance":"0","source_type":"APP","time":"2015-06-12T11:46:11+09:00","timestamp":1434077171244715915}

# Options

```
usage: firehose-to-syslog --api-endpoint=API-ENDPOINT --client-id=CLIENT-ID --client-secret=CLIENT-SECRET [<flags>]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --debug                        Enable debug mode. This disables forwarding to syslog
  --api-endpoint=API-ENDPOINT    Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io
  --doppler-endpoint=DOPPLER-ENDPOINT
                                 Overwrite default doppler endpoint return by /v2/info
  --syslog-server=SYSLOG-SERVER  Syslog server.
  --syslog-protocol="tcp"        Syslog protocol (tcp/udp/tcp+tls).
  --subscription-id="firehose"   Id for the subscription.
  --client-id=CLIENT-ID          Client ID.
  --client-secret=CLIENT-SECRET  Client secret.
  --skip-ssl-validation          Please don't
  --fh-keep-alive=25s            Keep Alive duration for the firehose consumer
  --min-retry-delay=500ms        Doppler Cloud Foundry Doppler min. retry delay duration
  --max-retry-delay=1m           Doppler Cloud Foundry Doppler max. retry delay duration
  --max-retry-count=1000         Doppler Cloud Foundry Doppler max. retry Count duration
  --logs-buffer-size=10000       Number of envelope to be buffered
  --events="LogMessage"          Comma separated list of events you would like. Valid options are ContainerMetric, CounterEvent, Error, HttpStartStop, LogMessage, ValueMetric
  --enable-stats-server          Will enable stats server on 8080
  --boltdb-path="my.db"          Bolt Database path
  --cc-pull-time=60s             CloudController Polling time in sec
  --extra-fields=""              Extra fields you want to annotate your events with, example: '--extra-fields=env:dev,something:other
  --mode-prof=""                 Enable profiling mode, one of [cpu, mem, block]
  --path-prof=""                 Set the Path to write profiling file
  --log-formatter-type=LOG-FORMATTER-TYPE
                                 Log formatter type to use. Valid options are text, json. If none provided, defaults to json.
  --cert-pem-syslog=""           Certificate Pem file
  --ignore-missing-apps          Enable throttling on cache lookup for missing apps
  --version                      Show application version.
```

** !!! **--events** Please use --help to get last updated event.


# TLS syslog endpoint.

Since v3 firehose-to-syslog support TLS syslog `--cert-pem-syslog` using PEM encoded cert file.
Please refer to https://github.com/RackSec/srslog/blob/master/script/gen-certs.py
for Cert generation.



# Event documentation

See the [dropsonde protocol documentation](https://github.com/cloudfoundry/dropsonde-protocol/tree/master/events) for details on what data is sent as part of each event.

# Caching
We use [boltdb](https://github.com/boltdb/bolt) for caching application name, org and space name.

We have 3 caching strategies:
* Pull all application data on start.
* Pull application data if not cached yet.
* Pull all application data every "cc-pull-time".

# To test and build


    # Setup repo
    go get github.com/cloudfoundry-community/firehose-to-syslog
    cd $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog

    # Test
	    ginkgo -r .

    # Build binary
    go build

# Deploy with Bosh

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Parsing the logs with Logstash

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)




# Profiling

To enable CPU Profiling you just need to add the profiling path ex ``` --mode-prof=cpu```

Run your program for some time and after that you can use the pprof tool
```bash
go tool pprof YOUR_EXECUTABLE cpu.pprof

(pprof) top 10
110ms of 110ms total (  100%)
Showing top 10 nodes out of 44 (cum >= 20ms)
      flat  flat%   sum%        cum   cum%
      30ms 27.27% 27.27%       30ms 27.27%  syscall.Syscall
      20ms 18.18% 45.45%       20ms 18.18%  ExternalCode
      20ms 18.18% 63.64%       20ms 18.18%  runtime.futex
      10ms  9.09% 72.73%       10ms  9.09%  adjustpointers
      10ms  9.09% 81.82%       10ms  9.09%  bytes.func·001
      10ms  9.09% 90.91%       20ms 18.18%  io/ioutil.readAll
      10ms  9.09%   100%       10ms  9.09%  runtime.epollwait
         0     0%   100%       60ms 54.55%  System
         0     0%   100%       20ms 18.18%  bufio.(*Reader).Read
         0     0%   100%       20ms 18.18%  bufio.(*Reader).fill
```

# Push as an App to Cloud Foundry

1. Create `doppler.firehose` enabled user or client
    - Use Client ID / Client Secret
      ```
      uaac target https://uaa.[your cf system domain]
      uaac token client get admin -s [your admin-secret]
      uaac client add firehose-to-syslog \
        --name firehose-to-syslog \
        --secret [your_client_secret] \
        --authorized_grant_types client_credentials,refresh_token \
        --authorities doppler.firehose,cloud_controller.global_auditor
      ```

1. Download the latest release of firehose-to-syslog.
    ```
    git clone https://github.com/cloudfoundry-community/firehose-to-syslog
    cd firehose-to-syslog
    ```

1. Utilize the CF cli to authenticate with your PCF instance.
    ```
    cf login -a https://api.[your cf system domain] -u [your id] --skip-ssl-validation
    ```
1. Push firehose-to-syslog.
    ```
    cf push firehose-to-syslog --no-start
    ```
1. Set environment variables with cf cli or in the [manifest.yml](./manifest.yml).
    ```
    cf set-env firehose-to-syslog API_ENDPOINT https://api.[your cf system domain]
    cf set-env firehose-to-syslog DOPPLER_ENDPOINT wss://doppler.[your cf system domain]:443
    cf set-env firehose-to-syslog SYSLOG_ENDPOINT [Your Syslog IP]:514
    cf set-env firehose-to-syslog LOG_EVENT_TOTALS true
    cf set-env firehose-to-syslog LOG_EVENT_TOTALS_TIME "10s"
    cf set-env firehose-to-syslog SKIP_SSL_VALIDATION true
    cf set-env firehose-to-syslog FIREHOSE_SUBSCRIPTION_ID firehose-to-syslog
    cf set-env firehose-to-syslog FIREHOSE_USER  [your doppler.firehose enabled user]
    cf set-env firehose-to-syslog FIREHOSE_PASSWORD  [your doppler.firehose enabled user password]
    cf set-env firehose-to-syslog FIREHOSE_CLIENT_ID  [your doppler.firehose enabled client id]
    cf set-env firehose-to-syslog FIREHOSE_CLIENT_SECRET  [your doppler.firehose enabled client secret]
    cf set-env firehose-to-syslog LOG_FORMATTER_TYPE [Log formatter type to use. Valid options are : text, json]
    ```
1. Turn off the health check if you're staging to Diego.
    ```
    cf set-health-check firehose-to-syslog process
    ```
1. Push the app.
    ```
    cf push firehose-to-syslog --no-route
    ```
