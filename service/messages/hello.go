package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type Hello struct {
	Address string
	Port    int
}

func (*Hello) GetType() MessageType { return HELLO }

func (h *Hello) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(&h)
}

func (h *Hello) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}
