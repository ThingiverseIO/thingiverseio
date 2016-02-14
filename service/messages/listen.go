package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type Listen struct {
	Function string
}

func (*Listen) GetType() MessageType { return LISTEN }

func (l *Listen) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(l)
}

func (l *Listen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}
