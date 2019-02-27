package firehoseclient_test

import (
	"code.cloudfoundry.org/go-loggregator"
	"code.cloudfoundry.org/go-loggregator/conversion"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"context"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehoseclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("V2adapter", func() {

	It("adapts an envelope stream to a channel of envelopes", func() {
		v2Env := &loggregator_v2.Envelope{
			Timestamp:  time.Now().Unix(),
			SourceId:   "test-source",
			InstanceId: "test-instance",
			Message: &loggregator_v2.Envelope_Log{
				Log: &loggregator_v2.Log{
					Payload: []byte("test-payload"),
				},
			},
		}

		stubStreamer := newStubStreamer()
		stubStreamer.envs = []*loggregator_v2.Envelope{v2Env}

		firehoseAdapter := firehoseclient.NewV2Adapter(stubStreamer)
		messages := firehoseAdapter.Firehose("test-subscription")

		expected := conversion.ToV1(v2Env)
		Eventually(messages).Should(Receive(Equal(expected[0])))
		Expect(stubStreamer.shardId).To(Equal("test-subscription"))

		Expect(stubStreamer.selectors).To(ConsistOf(
			&loggregator_v2.Selector{
				Message: &loggregator_v2.Selector_Log{
					Log: &loggregator_v2.LogSelector{},
				},
			},
			&loggregator_v2.Selector{
				Message: &loggregator_v2.Selector_Counter{
					Counter: &loggregator_v2.CounterSelector{},
				},
			},
			&loggregator_v2.Selector{
				Message: &loggregator_v2.Selector_Event{
					Event: &loggregator_v2.EventSelector{},
				},
			},
			&loggregator_v2.Selector{
				Message: &loggregator_v2.Selector_Gauge{
					Gauge: &loggregator_v2.GaugeSelector{},
				},
			},
			&loggregator_v2.Selector{
				Message: &loggregator_v2.Selector_Timer{
					Timer: &loggregator_v2.TimerSelector{},
				},
			},
		))

		Eventually(messages).Should(Receive(Equal(expected[0])))
	})

})

type stubStreamer struct {
	envs      []*loggregator_v2.Envelope
	shardId   string
	selectors []*loggregator_v2.Selector
}

func newStubStreamer() *stubStreamer {
	return &stubStreamer{}
}

func (s *stubStreamer) Stream(ctx context.Context, req *loggregator_v2.EgressBatchRequest) loggregator.EnvelopeStream {
	s.shardId = req.ShardId
	s.selectors = req.Selectors
	return func() []*loggregator_v2.Envelope {
		return s.envs
	}
}
