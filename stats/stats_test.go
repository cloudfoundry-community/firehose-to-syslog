// Taken from RakutenTech nozzle
// Thanks to them
package stats

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"testing"
)

func TestStatsInc(t *testing.T) {

	s := NewStats()

	loop := 20
	inc := 5

	var wg sync.WaitGroup
	wg.Add(loop)
	for i := 0; i < loop; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < inc; i++ {
				s.Inc(Consume)
			}
		}()
	}

	wg.Wait()

	expect := loop * inc
	if s.Consume != uint64(expect) {
		t.Fatalf("expect %d to be eq %d", s.Consume, expect)
	}
}

func TestStatsJson(t *testing.T) {
	s := NewStats()

	for i := 0; i < 100; i++ {
		s.Inc(Consume)
	}

	for i := 0; i < 50; i++ {
		s.Inc(PublishFail)
	}

	for i := 0; i < 50; i++ {
		s.Inc(Publish)
	}

	for i := 0; i < 100; i++ {
		s.Inc(SubInputBuffer)
	}

	for i := 0; i < 50; i++ {
		s.Dec(SubInputBuffer)
	}

	for i := 0; i < 100; i++ {
		s.Inc(Forwarded)
	}

	expect := `{
  "consume": 100,
  "consume_per_sec": 0,
  "consume_fail": 0,
  "consume_http_start_stop": 0,
  "consume_value_metric": 0,
  "consume_counter_event": 0,
  "consume_log_message": 0,
  "consume_error": 0,
  "consume_container_metric": 0,
  "consume_unknown": 0,
  "ignored": 0,
  "forwarded": 100,
  "publish": 50,
  "publish_per_sec": 0,
  "publish_fail": 50,
  "slow_consumer_alert": 0,
  "subinupt_buffer": 50,
  "delay": 0,
  "instance_id": 0
}`

	b, _ := s.Json()

	var buf bytes.Buffer
	json.Indent(&buf, b, "", "  ")
	if buf.String() != expect {
		t.Fatalf("expect %v to be eq %v", buf.String(), expect)
	}
}

func setEnv(k, v string) func() {
	prev := os.Getenv(k)
	os.Setenv(k, v)
	return func() {
		os.Setenv(k, prev)
	}
}

func TestNewStats(t *testing.T) {
	reset := setEnv(EnvCFInstanceIndex, "4")
	defer reset()

	stats := NewStats()
	if stats.InstanceID != 4 {
		t.Fatalf("expect %d to be eq 4", stats.InstanceID)
	}
}

func TestNewStats_nonNumber(t *testing.T) {
	reset := setEnv(EnvCFInstanceIndex, "ab")
	defer reset()

	stats := NewStats()
	if stats.InstanceID != defaultInstanceID {
		t.Fatalf("expect %d to be eq %d", stats.InstanceID, defaultInstanceID)
	}
}
