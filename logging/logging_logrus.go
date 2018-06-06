package logging

import (
	"fmt"
	"io/ioutil"
	"os"

	syslog "github.com/RackSec/srslog"
	logrus_syslog "github.com/shinji62/logrus-syslog-ng"
	"github.com/sirupsen/logrus"
)

type LoggingLogrus struct {
	Logger           *logrus.Logger
	syslogServer     string
	debugFlag        bool
	logFormatterType string
	certPath         string
	syslogProtocol   string
	skipSSL          bool
}

func NewLogging(SyslogServerFlag string, SysLogProtocolFlag string, LogFormatterFlag string, certP string, DebugFlag bool, skipSSLFlag bool) Logging {
	return &LoggingLogrus{
		Logger:           logrus.New(),
		syslogServer:     SyslogServerFlag,
		logFormatterType: LogFormatterFlag,
		syslogProtocol:   SysLogProtocolFlag,
		certPath:         certP,
		debugFlag:        DebugFlag,
		skipSSL:          skipSSLFlag,
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
		var hook logrus.Hook
		var err error
		if l.syslogProtocol == logrus_syslog.SecureProto {
			hook, err = logrus_syslog.NewSyslogHookTls(l.syslogServer, syslog.LOG_INFO, "doppler", l.certPath, l.skipSSL)

		} else {
			hook, err = logrus_syslog.NewSyslogHook(l.syslogProtocol, l.syslogServer, syslog.LOG_INFO, "doppler")
		}
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
