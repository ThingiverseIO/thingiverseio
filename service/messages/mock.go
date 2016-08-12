package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type Mock struct {
	Data interface{}
}

func (*Mock) New() Message{
	return new(Mock)
}

func (*Mock) GetType() MessageType { return MOCK }

func (m *Mock) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(&m)
}

func (m *Mock) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(m)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(&Mock{})
}
