FROM ubuntu:trusty
MAINTAINER Simon Johansson <simon.johansson@springer.com>

RUN apt-get update
RUN apt-get install -y ca-certificates
ADD dist/firehose-to-syslog-linux64 /

ENTRYPOINT ["/firehose-to-syslog-linux64"]
CMD ["--help"]
