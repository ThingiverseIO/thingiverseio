package messages

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type Hello struct {
	UUID    string
	Address string
	Port    int
}

func (*Hello) New() Message {
	return new(Hello)
}

func (*Hello) GetType() MessageType { return HELLO }

func (h *Hello) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(&h)
}

func (h *Hello) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(Hello))
}
