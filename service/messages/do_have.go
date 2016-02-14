package messages

import (
	"bytes"
	"strings"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type DoHave struct {
	Tag  string
}

func (*DoHave) GetType() MessageType { return DO_HAVE }

func (h *DoHave) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(&h)
}

func (h *DoHave) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}
