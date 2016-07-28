package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type DoHave struct {
	TagKey string
	TagValue string
}

func (*DoHave) New() Message{
	return new(DoHave)
}

func (*DoHave) GetType() MessageType { return DOHAVE }

func (h *DoHave) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(&h)
}

func (h *DoHave) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init(){
	registerMessage(new(DoHave))
}
