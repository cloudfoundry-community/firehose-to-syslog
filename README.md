This nifty util aggregates all the events from the firehose feature in
CloudFoundry.

To make it work you will need the following in your CF manifest

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
		--uaa-endpoint=https://uaa.domain.com/oauth/authorize \
		--doppler-endpoint=wss://doppler.domain.com \
		--firehose-user=firehoseuser \
		--firehose-password=firehosepassword \
		--all-events \
		--debug

	{"cf_app_id":"d6d2ad15-39e9-427f-bdde-e047f7989304","level":"info","message_type":"OUT","msg":"16:27:05 INFO  c.s.i.e.QueuedEmailService :: Starting queued mail processing","source_instance":"0","source_type":"App","time":"2014-12-16T17:27:05+01:00"}
	{"cf_app_id":"9f196e7c-133d-48a9-b905-4b3619e9126d","level":"info","message_type":"OUT","msg":"16:27:05 INFO  c.s.i.e.QueuedEmailService :: Starting queued mail processing","source_instance":"0","source_type":"App","time":"2014-12-16T17:27:05+01:00"}
	{"cf_app_id":"9f196e7c-133d-48a9-b905-4b3619e9126d","level":"info","message_type":"OUT","msg":"16:27:05 WARN  c.s.i.e.QueuedEmailService :: Cannot process mail as there is a lock in place","source_instance":"1","source_type":"App","time":"2014-12-16T17:27:05+01:00"}
	{"cf_app_id":"cf72f41b-f0e3-40dc-8c10-1b45262bd1f8","level":"info","message_type":"OUT","msg":"wakawakwaka.domain.com - [16/12/2014:16:27:05 +0000] \"GET /internal/status HTTP/1.1\" 200 6 \"-\" \"-\" xx.yy.zz.yy:36146 x_forwarded_for:\"xx.yy.zz.qq\" vcap_request_id:547ce74f-226a-44cc-4f69-9d41e75fe77a response_time:0.004542139 app_id:cf72f41b-f0e3-40dc-8c10-1b45262bd1f8\n","source_instance":"0","source_type":"RTR","time":"2014-12-16T17:27:05+01:00"}

# To build

    # Setup repo
    go get github.com/SpringerPE/firehose-to-syslog
    cd $GOPATH/src/github.com/SpringerPE/firehose-to-syslog

    # Build binary
    godep go build

# Deploy with Bosh

[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)

# Run with Docker

	sudo docker run --restart=always -d springerplatformengineering/firehose-to-logstash --uaa-endpoint=https://uaa.10.244.0.34.xip.io/oauth/authorize --doppler-endpoint=wss://doppler.10.244.0.34.xip.io:443 --firehose-user=admin --firehose-password=admin --subscription-id=firehose-a --debug --skip-ssl-validation --all-events

# Run agains a bosh-lite CF deployment

    godep go run main.go \
    --uaa-endpoint=https://uaa.10.244.0.34.xip.io/oauth/authorize \
    --doppler-endpoint=wss://doppler.10.244.0.34.xip.io:443 \
    --firehose-user=admin \
    --firehose-password=admin \
    --subscription-id=firehose-a \
    --debug \
    --all-events \
    --skip-ssl-validation

# Parsing the logs with Logstash

There is a grok-pattern folder with a couple of filters for app
and and routing logs. But I would strongy encourage to use
[logsearch-for-cloudfoundry](https://github.com/logsearch/logsearch-for-cloudfoundry)
that provides >= functionality but in a nicer package.

# Devel

This is a
[Git Flow](http://nvie.com/posts/a-successful-git-branching-model/)
project. Please fork and branch your features from develop.

# Contributors

* [Ed King](https://github.com/teddyking) - Added support to skip ssl
validation.
* [Mark Alston](https://github.com/malston) - Added support for more
  events and general code cleaup.
