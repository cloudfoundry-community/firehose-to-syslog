package logging

import (
	"fmt"
	"os"
	"time"
)

//go:generate counterfeiter . Logging

type Logging interface {
	Connect() bool
	ShipEvents(map[string]interface{}, string)
}

func LogStd(message string, force bool) {
	Log(message, force, false, nil)
}

func LogError(message string, errMsg interface{}) {
	Log(message, false, true, errMsg)
}

func Log(message string, force bool, isError bool, err interface{}) {

	if force || isError {

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
