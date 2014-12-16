FROM debian
MAINTAINER Simon Johansson <simon.johansson@springer.com>

RUN apt-get update && apt-get install -y ca-certificates

ADD . code
WORKDIR code
ENTRYPOINT ./run.sh
