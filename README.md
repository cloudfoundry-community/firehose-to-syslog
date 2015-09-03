# Firehose-to-syslog

This nifty util aggregates all the events from the firehose feature in
CloudFoundry.

	./firehose-to-syslog \
              --api-address="https://api.10.244.0.34.xip.io" \
              --skip-ssl-validation \
              --debug
	....
	....
	{"cf_app_id":"c5cb762b-b7bb-44b6-97d1-2b612d4baba9","cf_app_name":"lattice","cf_org_id":"fb5777e6-e234-4832-8844-773114b505b0","cf_org_name":"GWENN","cf_origin":"firehose","cf_space_id":"3c910823-22e7-41ff-98de-094759594398","cf_space_name":"GWENN-SPACE","event_type":"LogMessage","level":"info","message_type":"OUT","msg":"Lattice-app. Says Hello. on index: 0","origin":"rep","source_instance":"0","source_type":"APP","time":"2015-06-12T11:46:11+09:00","timestamp":1434077171244715915}

# Options

```
usage: firehose-to-syslog --api-endpoint=API-ENDPOINT [<flags>]

Flags:
  --help              Show help (also see --help-long and --help-man).
  --debug             Enable debug mode. This disables forwarding to syslog
  --api-endpoint=API-ENDPOINT  
                      Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io
  --doppler-endpoint=DOPPLER-ENDPOINT  
                      Overwrite default doppler endpoint return by /v2/info
  --syslog-server=SYSLOG-SERVER  
                      Syslog server.
  --subscription-id="firehose"  
                      Id for the subscription.
  --user="admin"      Admin user.
  --password="admin"  Admin password.
  --skip-ssl-validation  
                      Please don't
  --events="LogMessage"  
                      Comma seperated list of events you would like. Valid options are LogMessage, ValueMetric, CounterEvent, Error, ContainerMetric, Heartbeat, HttpStart,
                      HttpStop, HttpStartStop
  --boltdb-path="my.db"  
                      Bolt Database path
  --cc-pull-time=60s  CloudController Pooling time in sec
  --version           Show application version.
  --mode-prof         Enable profiling mode, one of [cpu, mem, block]
  --path-prof         Set the Path to write Profiling file
	--filters=FILTERS   Pipe seperated filtering for messages. Possible keys: org_name, org_id, space_name, space_id, app_name, app_id. Values are comma seperated. If set, works as a
										  whitelist filter. Example: --filters="org_name:org1,org2|space_id:asff-12ffa,1122-dbfa-aaaa|app_name:app1"

```

#Endpoint definition

We use [gocf-client](https://github.com/cloudfoundry-community/go-cfclient) which will call the CF endpoint /v2/info to get Auth., doppler endpoint.

But for doppler endpoint you can overwrite it with ``` --doppler-address ``` as we know some people use different endpoint.

# Event documentation

See the [dropsonde protocol documentation](https://github.com/cloudfoundry/dropsonde-protocol/tree/master/events) for details on what data is sent as part of each event.

# Caching
We use [boltdb](https://github.com/boltdb/bolt) for caching application name, org and space name.

We have 3 caching strategies:
* Pull all application data on start.
* Pull application data if not cached yet.
* Pull all application data every "cc-pull-time".

# Filters

It is possible to filter messages (works only for ``` LogMessage ```) by organization name or Id, space name or Id, application name or Id.
The filter, if applied, works as a whitelist of desired orgs, spaces and apps that we want to get messages from.

# To test and build


    # Setup repo
    go get github.com/eljuanchosf/firehose-to-syslog
    cd $GOPATH/src/github.com/eljuanchosf/firehose-to-syslog

    # Test
	  ginkgo -r .

    # Build binary
    godep go build

# Deploy with Bosh

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Run agains a bosh-lite CF deployment

    godep go run main.go \
		--debug \
		--skip-ssl-validation \
		--api-address="https://api.10.244.0.34.xip.io"

# Parsing the logs with Logstash

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Devel

This is a
[Git Flow](http://nvie.com/posts/a-successful-git-branching-model/)
project. Please fork and branch your features from develop.

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

For Mac OSX golang profiling do not work.

# Contributors

* [Ed King](https://github.com/teddyking) - Added support to skip ssl
validation.
* [Mark Alston](https://github.com/malston) - Added support for more
  events and general code cleaup.
* [Etourneau Gwenn](https://github.com/shinji62) - Added validation of
  selected events and general code cleanup, caching system..
