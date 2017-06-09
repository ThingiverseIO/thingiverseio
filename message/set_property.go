package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type SetProperty struct {
	Name  string
	Value []byte
}

func (*SetProperty) New() Message {
	return new(SetProperty)
}

func (*SetProperty) GetType() Type { return SETPROPERTY }

func (m *SetProperty) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(m)
}

func (m *SetProperty) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(SetProperty))
}
