package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type GetProperty struct {
	Name string
}

func (*GetProperty) New() Message {
	return new(GetProperty)
}

func (*GetProperty) GetType() Type { return GETPROPERTY }

func (m *GetProperty) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *GetProperty) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(GetProperty))
}
