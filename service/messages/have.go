package messages

import (
	"bytes"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type Have struct {
	Have bool
	Tag  string
}

func (*Have) GetType() MessageType { return HAVE }

func (h *Have) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(&h)
}

func (h *Have) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}
