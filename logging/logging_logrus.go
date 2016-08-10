package logging

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"io/ioutil"
	"log/syslog"
	"os"
)

type LoggingLogrus struct {
	Logger           *logrus.Logger
	syslogServer     string
	debugFlag        bool
	logFormatterType string
	syslogProtocol   string
}

func NewLogging(SyslogServerFlag string, SysLogProtocolFlag string, LogFormatterFlag string, DebugFlag bool) Logging {
	return &LoggingLogrus{
		Logger:           logrus.New(),
		syslogServer:     SyslogServerFlag,
		logFormatterType: LogFormatterFlag,
		syslogProtocol:   SysLogProtocolFlag,
		debugFlag:        DebugFlag,
	}
}

func (l *LoggingLogrus) Connect() bool {

	success := false
	l.Logger.Formatter = GetLogFormatter(l.logFormatterType)

	if !l.debugFlag {
		l.Logger.Out = ioutil.Discard
	} else {
		l.Logger.Out = os.Stdout
	}

	if l.syslogServer != "" {
		hook, err := logrus_syslog.NewSyslogHook(l.syslogProtocol, l.syslogServer, syslog.LOG_INFO, "doppler")
		if err != nil {
			LogError(fmt.Sprintf("Unable to connect to syslog server [%s]!\n", l.syslogServer), err.Error())
		} else {
			LogStd(fmt.Sprintf("Received hook to syslog server [%s]!\n", l.syslogServer), false)
			l.Logger.Hooks.Add(hook)
			success = true
		}
	}
	return success
}

func (l *LoggingLogrus) ShipEvents(eventFields map[string]interface{}, Message string) {
	l.Logger.WithFields(eventFields).Info(Message)
}

func GetLogFormatter(logFormatterType string) logrus.Formatter {
	switch logFormatterType {
	case "text":
		return &logrus.TextFormatter{}
	default:
		return &logrus.JSONFormatter{}
	}
}
