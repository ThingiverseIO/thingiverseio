package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type Listen struct {
	Function string
}

func (*Listen) New() Message {
	return new(Listen)
}

func (*Listen) GetType() Type { return LISTEN }

func (l *Listen) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(l)
}

func (l *Listen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(Listen))
}
