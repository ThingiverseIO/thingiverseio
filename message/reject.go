package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

// Reject message indicates that a sent request have been rejected.
type Reject struct {
	UUID   string
	Output string
	Reason string
}

func (*Reject) GetType() Type { return REJECT }

func (*Reject) New() Message {
	return new(Reject)
}

func (h *Reject) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(&h)
}

func (h *Reject) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(Reject))
}
