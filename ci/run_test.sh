#!/bin/bash
cd firehose-to-syslog-source/
mkdir -p /go/src/github.com/cloudfoundry-community/firehose-to-syslog/
cp -aR . /go/src/github.com/cloudfoundry-community/firehose-to-syslog/
cd /go/src/github.com/cloudfoundry-community/firehose-to-syslog/
go get github.com/tools/godep
godep restore
go get github.com/onsi/ginkgo/ginkgo
ginkgo -r
