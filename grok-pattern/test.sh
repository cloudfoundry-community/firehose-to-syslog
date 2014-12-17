#!/usr/bin/env bash
set -e

echo "Unit testing grok patterns"
docker run -v $PWD:/tests springerplatformengineering/logstash-grok-tester:1.4.0
