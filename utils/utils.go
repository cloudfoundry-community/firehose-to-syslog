package utils

import (
	"encoding/binary"
	"fmt"
	"github.com/cloudfoundry/sonde-go/events"
	"strings"
)

func FormatUUID(uuid *events.UUID) string {
	if uuid == nil {
		return ""
	}
	var uuidBytes [16]byte
	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}

func ConcatFormat(stringList []string) string {
	r := strings.NewReplacer(".", "_")
	for i, s := range stringList {
		stringList[i] = strings.TrimSpace(r.Replace(s))
	}

	return strings.Join(stringList, ".")
}
