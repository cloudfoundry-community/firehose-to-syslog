package utils

import (
  "fmt"
  "strings"
)

func GetKeyValueFromString(kvPair string) (string, string, error) {
	values := strings.Split(kvPair, ":")
	if len(values) != 2 {
		return "", "", fmt.Errorf("When splitting %s by ':' there must be exactly 2 values, got these values %s", kvPair, values)
	}
	return strings.TrimSpace(values[0]), strings.TrimSpace(values[1]), nil
}
