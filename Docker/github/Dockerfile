FROM google/golang:1.4
MAINTAINER Gwenn Etourneau <gwenn.etourneau@gmail.com>

RUN go get github.com/tools/godep
RUN go get github.com/cloudfoundry-community/firehose-to-syslog
RUN cd $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog  \
    ; CGO_ENABLED=0 godep go build  -a --installsuffix cgo --ldflags="-s"
RUN cp $GOPATH/src/github.com/cloudfoundry-community/firehose-to-syslog/firehose-to-syslog /gopath/bin/
COPY Dockerfile.final /gopath/bin/Dockerfile
RUN ls -lah /gopath/bin/

CMD docker build  -t getourneau/firehose-to-syslog /gopath/bin
