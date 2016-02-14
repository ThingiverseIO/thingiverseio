package messages

import (
	"bytes"
	"gopkg.in/vmihailenco/msgpack.v2"
	"strings"
)

type HelloOk struct {
}

func (*HelloOk) GetType() MessageType { return HELLO_OK }

func (h *HelloOk) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(&h)
}

func (h *HelloOk) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}
