
package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StopConsume struct {
	Stream string
}

func (*StopConsume) New() Message {
	return new(StopConsume)
}

func (*StopConsume) GetType() Type { return STOPCONSUME }

func (m *StopConsume) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *StopConsume) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StopConsume))
}
