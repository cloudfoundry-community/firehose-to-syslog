This nifty util aggregates all the events from the firehose feature in
CloudFoundry.

To make it work unless you want to run with the admin user, you will need the following in your CF manifest.

```
	uaa:
		clients:
			cf:
				scope: '....,doppler.firehose'
	scim:
		users:
			- firehoseuser|firehosepassword|doppler.firehose

```

Then you should be able to do this and get some nice logs.

	./firehose-to-logstash \
		--domain=cf.installation.domain.com \
		--user=username \
		--password=password \
		--debug

	{"cf_app_id":"e626413b-f1f8-436d-8963-c46f7cb345eb","cf_app_name":"php-diego-one","cf_org_id":"ebd95a83-5b6a-43ff-af67-a234ece3fb78","cf_org_name":"GWENN","cf_space_id":"a2e4c75b-fe02-4078-abfb-87539352aeac","cf_space_name":"GWENN-SPACE","cpu_percentage":1.4523587130944957,"disk_bytes":0,"event_type":"ContainerMetric","instance_index":0,"level":"info","memory_bytes":14110720,"msg":"","origin":"executor","time":"2015-04-17T13:59:52-07:00"}

# Options

```
usage: firehose-to-syslog [<flags>]

Flags:
  --help              Show help.
  --debug             Enable debug mode. This disables forwarding to syslog
  --domain="10.244.0.34.xip.io" Domain of your CF installation.
  --syslog-server=SYSLOG-SERVER
                      Syslog server.
  --subscription-id=firehose  Id for the subscription.
  --user=admin      Admin user.
  --password=admin  Admin password.
  --skip-ssl-validation Please don\'t
  --events=LogMessage Comma seperated list of events you would like. Valid options are HttpStart, HttpStop, Heartbeat, HttpStartStop, LogMessage, ValueMetric, CounterEvent, Error, ContainerMetric
  --boltdb-path='my.db' Bolt Database path
  --cc-pull-time=60s  CloudController Pooling time in sec
  --version           Show application version.
```
# Event documentation

See the [dropsonde protocol documentation](https://github.com/cloudfoundry/dropsonde-protocol/tree/master/events) for details on what data is sent as part of each event.

# Caching
We use [boltdb](https://github.com/boltdb/bolt) for caching application name, org and space name.

We have 2 caching strategies:
* Pull all application data on start
* Pull by application id if not cached yet
* Pull every "cc-pull-time" all applications data

# To build


    # Setup repo
    go get github.com/cloudfoundry-community/firehose-to-syslog
    cd $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog

    # Build binary
    godep go build

# Deploy with Bosh

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Run agains a bosh-lite CF deployment

    godep go run main.go \
		--debug \
		--skip-ssl-validation

# Parsing the logs with Logstash

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Devel

This is a
[Git Flow](http://nvie.com/posts/a-successful-git-branching-model/)
project. Please fork and branch your features from develop.

# Contributors

* [Ed King](https://github.com/teddyking) - Added support to skip ssl
validation.
* [Mark Alston](https://github.com/malston) - Added support for more
  events and general code cleaup.
* [Etourneau Gwenn](https://github.com/shinji62) - Added validation of
  selected events and general code cleanup, caching system..
