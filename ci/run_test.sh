#!/bin/bash
cd firehose-to-syslog-source/
go get github.com/tools/godep
godep restore
go get github.com/onsi/ginkgo/ginkgo
ginkgo -r
