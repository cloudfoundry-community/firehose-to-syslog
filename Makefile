all: lint vet test linux32 linux64 darwin64

lint:
	for SRC_FILE in `find . -type f -not -path "./Godeps/*" -name "*.go"`; do golint $$SRC_FILE; done

vet:
	find . -type f -not -path "./Godeps/*" -name "*.go"  | xargs go tool vet -v
test:
	ginkgo -r .

linux32:
	GOARCH=386 GOOS=linux godep go build -o firehose-to-syslog-linux32

linux64:
	GOARCH=amd64 GOOS=linux godep go build -o firehose-to-syslog-linux64

darwin64:
	GOARCH=amd64 GOOS=darwin godep go build -o firehose-to-syslog-darwin64

clean:
	$(RM) firehose-to-syslog-linux32
	$(RM) firehose-to-syslog-linux64
	$(RM) firehose-to-syslog-darwin64
