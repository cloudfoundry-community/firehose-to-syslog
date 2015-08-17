#!/bin/bash
DIR=`dirname "$(readlink -f "$0")"` 
pushd $DIR
  docker build -t cloudfoundry-community/firehose-syslog-build ./github/
  docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):$(which docker) -ti --name firehose-syslog-build cloudfoundry-community/firehose-syslog-build
  docker rm firehose-syslog-build
  docker rmi cloudfoundry-community/firehose-syslog-build
popd
