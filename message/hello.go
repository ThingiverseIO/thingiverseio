package message

import (
	"bytes"

	"github.com/ThingiverseIO/thingiverseio/descriptor"
	"github.com/ugorji/go/codec"
)

type Hello struct {
	NetworkDetails [][]byte
	Tag            descriptor.Tag
}

func (h *Hello) New() Message{
	return new(Hello)
}

func (Hello) GetType() Type { return HELLO }

func (h *Hello) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(h)
}

func (h Hello) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(h)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(Hello))
}
