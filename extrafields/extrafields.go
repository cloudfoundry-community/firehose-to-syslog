package extrafields

import (
	"github.com/eljuanchosf/firehose-to-syslog/utils"
	"strings"
)

func ParseExtraFields(extraEventsString string) (map[string]string, error) {
	extraEvents := map[string]string{}

	for _, kvPair := range strings.Split(extraEventsString, ",") {
		if kvPair != "" {
			cleaned := strings.TrimSpace(kvPair)
			k, v, err := utils.GetKeyValueFromString(cleaned)
			if err != nil {
				return nil, err
			}
			extraEvents[k] = v
		}
	}
	return extraEvents, nil
}
