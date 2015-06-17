all: test linux32 linux64 darwin64

test:
	ginkgo -r .

linux32:
	GOARCH=386 GOOS=linux godep go build -o dist/firehose-to-syslog-linux32

linux64:
	GOARCH=amd64 GOOS=linux godep go build -o dist/firehose-to-syslog-linux64

darwin64:
	GOARCH=amd64 GOOS=darwin godep go build -o dist/firehose-to-syslog-darwin64

clean:
	$(RM) dist/*
