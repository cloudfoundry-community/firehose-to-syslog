// Taken from RakutenTech nozzle
// Thanks to them

package stats

import (
	"os"
	"strconv"
	"sync/atomic"
	"time"

	json "github.com/mailru/easyjson"
)

const (
	defaultInstanceID int = 0
)

const (
	EnvCFInstanceIndex = "CF_INSTANCE_INDEX"
)

type StatsType int

const (
	Consume                StatsType = iota // messages received
	ConsumeFail                             // messages failed to be consumed
	ConsumeHttpStartStop                    // HttpStartStop messages received
	ConsumeValueMetric                      // ValueMetric messages received
	ConsumeCounterEvent                     // CounterEvent messages received
	ConsumeLogMessage                       // LogMessage messages received
	ConsumeError                            // Error messages received
	ConsumeContainerMetric                  // ContainerMetric messages received
	ConsumeUnknown                          // unknown type messages received
	Ignored                                 // messages dropped because of no forwarding rule
	Forwarded                               // messages
	Publish                                 // messages
	PublishFail                             // messages
	SlowConsumerAlert                       // slow consumer alerts emitted by noaa?
	SubInputBuffer                          // messages in buffer
)

// Stats stores various stats infomation
//easyjson:json
type Stats struct {
	Consume                uint64 `json:"consume"`
	ConsumePerSec          uint64 `json:"consume_per_sec"`
	ConsumeFail            uint64 `json:"consume_fail"`
	ConsumeHttpStartStop   uint64 `json:"consume_http_start_stop"`
	ConsumeValueMetric     uint64 `json:"consume_value_metric"`
	ConsumeCounterEvent    uint64 `json:"consume_counter_event"`
	ConsumeLogMessage      uint64 `json:"consume_log_message"`
	ConsumeError           uint64 `json:"consume_error"`
	ConsumeContainerMetric uint64 `json:"consume_container_metric"`
	ConsumeUnknown         uint64 `json:"consume_unknown"`
	Ignored                uint64 `json:"ignored"`
	Forwarded              uint64 `json:"forwarded"`

	Publish       uint64 `json:"publish"`
	PublishPerSec uint64 `json:"publish_per_sec"`

	// This is same as the number of dropped message
	PublishFail uint64 `json:"publish_fail"`

	SlowConsumerAlert uint64 `json:"slow_consumer_alert"`

	// SubInputBuffer is used to count number of current
	// buffer on subInput.
	SubInputBuffer int64 `json:"subinupt_buffer"`

	// Delay is Forwarded - (Publish + PublishFail)
	// This indicate how slow publish to kafka
	//Delay uint64 `json:"delay"`

	// InstanceID is ID for nozzle instance.
	// This is used to identify stats from different instances.
	// By default, it's defaultInstanceID
	InstanceID int `json:"instance_id"`
}

func NewStats() *Stats {
	instanceID := defaultInstanceID
	if idStr := os.Getenv(EnvCFInstanceIndex); len(idStr) != 0 {
		var err error
		instanceID, err = strconv.Atoi(idStr)
		if err != nil {
			// If it's failed to conv str to int
			// use default var
			instanceID = defaultInstanceID
		}
	}

	return &Stats{
		InstanceID: instanceID,
	}
}

func (s *Stats) Json() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Stats) PerSec() {
	lastConsume, lastPublish := uint64(0), uint64(0)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			s.ConsumePerSec = atomic.LoadUint64(&s.Consume) - atomic.LoadUint64(&lastConsume)
			s.PublishPerSec = atomic.LoadUint64(&s.Publish) - atomic.LoadUint64(&lastPublish)

			lastConsume = atomic.LoadUint64(&s.Consume)
			lastPublish = atomic.LoadUint64(&s.Publish)
		}
	}
}

func (s *Stats) Inc(statsType StatsType) {
	switch statsType {
	case Consume:
		atomic.AddUint64(&s.Consume, 1)
	case ConsumeFail:
		atomic.AddUint64(&s.ConsumeFail, 1)
	case Publish:
		atomic.AddUint64(&s.Publish, 1)
	case PublishFail:
		atomic.AddUint64(&s.PublishFail, 1)
	case SlowConsumerAlert:
		atomic.AddUint64(&s.SlowConsumerAlert, 1)
	case ConsumeHttpStartStop:
		atomic.AddUint64(&s.ConsumeHttpStartStop, 1)
	case ConsumeValueMetric:
		atomic.AddUint64(&s.ConsumeValueMetric, 1)
	case ConsumeCounterEvent:
		atomic.AddUint64(&s.ConsumeCounterEvent, 1)
	case ConsumeLogMessage:
		atomic.AddUint64(&s.ConsumeLogMessage, 1)
	case ConsumeError:
		atomic.AddUint64(&s.ConsumeError, 1)
	case ConsumeContainerMetric:
		atomic.AddUint64(&s.ConsumeContainerMetric, 1)
	case ConsumeUnknown:
		atomic.AddUint64(&s.ConsumeUnknown, 1)
	case Ignored:
		atomic.AddUint64(&s.Ignored, 1)
	case Forwarded:
		atomic.AddUint64(&s.Forwarded, 1)
	case SubInputBuffer:
		atomic.AddInt64(&s.SubInputBuffer, 1)
	}
}

func (s *Stats) Dec(statsType StatsType) {
	switch statsType {
	case SubInputBuffer:
		atomic.AddInt64(&s.SubInputBuffer, -1)
	}
}
