package messages

import (
	"bytes"
	"strings"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Hello struct {
	Address string
	Port    int
}

func (*Hello) GetType() MessageType { return HELLO }

func (h *Hello) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(&h)
}

func (h *Hello) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}
