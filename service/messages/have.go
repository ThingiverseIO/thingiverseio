package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type Have struct {
	Have bool
	TagKey string
	TagValue string
}

func (*Have) New() Message{
	return new(Have)
}

func (*Have) GetType() MessageType { return HAVE }

func (h *Have) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(&h)
}

func (h *Have) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init(){
	registerMessage(new(Have))
}
