package extrafields

import (
	"fmt"
	"strings"
)

func getKeyValueFromString(kvPair string) (string, string, error) {
	values := strings.Split(kvPair, ":")
	if len(values) != 2 {
		return "", "", fmt.Errorf("When splitting %s by ':' there must be exactly 2 values, got these values %s", kvPair, values)
	}
	return strings.TrimSpace(values[0]), strings.TrimSpace(values[1]), nil
}

func ParseExtraFields(extraEventsString string) (map[string]string, error) {
	extraEvents := map[string]string{}

	for _, kvPair := range strings.Split(extraEventsString, ",") {
		if kvPair != "" {
			cleaned := strings.TrimSpace(kvPair)
			k, v, err := getKeyValueFromString(cleaned)
			if err != nil {
				return nil, err
			}
			extraEvents[k] = v
		}
	}
	return extraEvents, nil
}

func FieldExist(fieldList map[string]string, field string) bool {
	_, presence := fieldList[field]
	return presence

}
