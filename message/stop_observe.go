
package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StopObserve struct {
	Property string
}

func (*StopObserve) New() Message {
	return new(StopObserve)
}

func (*StopObserve) GetType() Type { return STOPOBSERVE }

func (m *StopObserve) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *StopObserve) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StopObserve))
}
