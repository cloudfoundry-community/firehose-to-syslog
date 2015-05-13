package logging

import (
	"github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"io/ioutil"
	"log/syslog"
	"os"
)

func SetupLogging(syslogServer string, debug bool) {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	if !debug {
		logrus.SetOutput(ioutil.Discard)
	}
	if syslogServer != "" {
		hook, err := logrus_syslog.NewSyslogHook("tcp", syslogServer, syslog.LOG_INFO, "doppler")
		if err != nil {
			logrus.Error("Unable to connect to syslog server.")
		} else {
			logrus.AddHook(hook)
		}
	}
}
