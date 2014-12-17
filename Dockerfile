FROM golang:1.4.0-onbuild
MAINTAINER Simon Johansson <simon.johansson@springer.com>


ENTRYPOINT ["/go/bin/app"]
CMD ["--help"]
