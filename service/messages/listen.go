package messages

import (
	"bytes"
	"strings"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Listen struct {
	Function string
}

func (*Listen) GetType() MessageType { return LISTEN }

func (l *Listen) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(l)
}

func (l *Listen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}
