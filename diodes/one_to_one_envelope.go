package diodes

import (
	gendiodes "code.cloudfoundry.org/diodes"
	"github.com/cloudfoundry/sonde-go/events"
)

// ManyToOneEnvelope diode is optimal for many writers and a single reader for
// V1 envelopes.
type OneToOneEnvelope struct {
	d *gendiodes.Poller
}

// NewManyToOneEnvelope returns a new ManyToOneEnvelope diode to be used with
// many writers and a single reader.
func NewOneToOneEnvelope(size int, alerter gendiodes.Alerter) *OneToOneEnvelope {
	return &OneToOneEnvelope{
		d: gendiodes.NewPoller(gendiodes.NewManyToOne(size, alerter)),
	}
}

// Set inserts the given V1 envelope into the diode.
func (d *OneToOneEnvelope) Set(data *events.Envelope) {
	d.d.Set(gendiodes.GenericDataType(data))
}

// TryNext returns the next V1 envelope to be read from the diode. If the
// diode is empty it will return a nil envelope and false for the bool.
func (d *OneToOneEnvelope) TryNext() (*events.Envelope, bool) {
	data, ok := d.d.TryNext()
	if !ok {
		return nil, ok
	}

	return (*events.Envelope)(data), true
}

// Next will return the next V1 envelope to be read from the diode. If the
// diode is empty this method will block until anenvelope is available to be
// read.
func (d *OneToOneEnvelope) Next() *events.Envelope {
	data := d.d.Next()
	return (*events.Envelope)(data)
}
