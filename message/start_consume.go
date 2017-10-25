package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StartConsume struct {
	Stream string
}

func (*StartConsume) New() Message {
	return new(StartConsume)
}

func (*StartConsume) GetType() Type { return STARTCONSUME }

func (m *StartConsume) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *StartConsume) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StartConsume))
}
