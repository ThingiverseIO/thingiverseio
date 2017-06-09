package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StartObserve struct {
	Property string
}

func (*StartObserve) New() Message {
	return new(StartObserve)
}

func (*StartObserve) GetType() Type { return STARTOBSERVE }

func (m *StartObserve) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *StartObserve) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StartObserve))
}
