all: test compile
compile: linux32 linux64 darwin64 


test:
	ginkgo -r -v .

linux32:
	 CGO_ENABLED=0 GOARCH=386 GOOS=linux go build --ldflags="-X main.version=${BUILD_NUMBER}" -o dist/linux/386/firehose-to-syslog_linux_386

linux64:
	 CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build --ldflags="-X main.version=${BUILD_NUMBER}" -o dist/linux/amd64/firehose-to-syslog_linux_amd64

darwin64:
	GOARCH=amd64 GOOS=darwin go build --ldflags="-X main.version=${BUILD_NUMBER}" -o dist/darwin/amd64/firehose-to-syslog_darwin_amd64

clean:
	-rm -rf dist/*
	-rm -rf *.prof

docker-dev:
	$(SHELL) ./Docker/build-dev.sh

docker-final:
	$(SHELL) ./Docker/build.sh     