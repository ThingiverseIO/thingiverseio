package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type HelloOk struct {
	Have bool
}

func (h *HelloOk) New() Message {
	return new(HelloOk)
}

func (*HelloOk) GetType() Type { return HELLOOK }

func (h *HelloOk) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(h)
}

func (h *HelloOk) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(HelloOk))
}
