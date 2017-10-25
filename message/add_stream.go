package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type AddStream struct {
	Name  string
	Value []byte
}

func (*AddStream) New() Message {
	return new(AddStream)
}

func (*AddStream) GetType() Type { return ADDSTREAM }

func (m *AddStream) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *AddStream) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(AddStream))
}
