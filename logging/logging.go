package logging

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"io/ioutil"
	"log/syslog"
	"os"
	"time"
)

var (
	debugFlag    bool
	syslogServer string
)

func Connect() bool {

	success := false

	logrus.SetFormatter(&logrus.JSONFormatter{})

	if !debugFlag {
		logrus.SetOutput(ioutil.Discard)
	} else {
		logrus.SetOutput(os.Stdout)
	}
	if syslogServer != "" {
		hook, err := logrus_syslog.NewSyslogHook("tcp", syslogServer, syslog.LOG_INFO, "doppler")
		if err != nil {
			LogError(fmt.Sprintf("Unable to connect to syslog server [%s]!\n", syslogServer), err.Error())
		} else {
			LogStd(fmt.Sprintf("Received hook to syslog server [%s]!\n", syslogServer), false)
			logrus.AddHook(hook)

			success = true
		}
	}

	return success
}

func SetupLogging(syslogSvr string, debug bool) {
	debugFlag = debug
	syslogServer = syslogSvr
}

func LogStd(message string, force bool) {
	Log(message, force, false, nil)
}

func LogError(message string, errMsg interface{}) {

	Log(message, false, true, errMsg)
}

func Log(message string, force bool, isError bool, err interface{}) {

	if debugFlag || force || isError {

		writer := os.Stdout
		var formattedMessage string

		if isError {
			writer = os.Stderr
			formattedMessage = fmt.Sprintf("[%s] Exception occurred! Message: %s Details: %v", time.Now().String(), message, err)
		} else {
			formattedMessage = fmt.Sprintf("[%s] %s", time.Now().String(), message)
		}

		fmt.Fprintln(writer, formattedMessage)
	}
}
