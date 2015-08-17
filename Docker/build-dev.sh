#!/bin/bash
DIR=`dirname "$(readlink -f "$0")"`
DEV_MAPPER=""
[ -f /usr/lib/libdevmapper.so.1.02 ] && DEV_MAPPER="-v /usr/lib/libdevmapper.so.1.02:/usr/lib/libdevmapper.so.1.02"
 

pushd $DIR

  pushd ../
    tar -zcvf firehose-to-syslog.tgz --exclude="Docker*" \
        --exclude=".git" --exclude="my.db" --exclude="dist" \
        --exclude="firehose-to-syslog"  ./  
  popd
  mv ../firehose-to-syslog.tgz ./dev/
  docker build -t cloudfoundry-community/firehose-to-syslog-build-dev $(PWD)/dev/
  docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):$(which docker) $DEV_MAPPER  -ti --name firehose-to-syslog-build-dev cloudfoundry-community/firehose-to-syslog-build-dev
  rm dev/firehose-to-syslog.tgz

popd

docker rm firehose-to-syslog-build-dev
docker rmi cloudfoundry-community/firehose-to-syslog-build-dev
