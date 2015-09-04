FROM google/golang:1.4
MAINTAINER Gwenn Etourneau <gwenn.etourneau@gmail.com>

RUN go get github.com/tools/godep
ADD firehose-to-syslog.tgz   $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog/
RUN cd $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog \
    ; CGO_ENABLED=0 godep go build  -a --installsuffix cgo --ldflags="-s"
RUN cd $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog 
RUN cp $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog/firehose-to-syslog /gopath/bin/
COPY Dockerfile.final /gopath/bin/Dockerfile

CMD docker build  --no-cache -t getourneau/firehose-to-syslog-dev /gopath/bin/
