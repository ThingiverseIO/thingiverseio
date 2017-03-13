package thingiverseio

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func encode(data interface{}) (encoded []byte, err error) {
	var buf bytes.Buffer
	enc := codec.NewEncoder(&buf, &mh)
	if err = enc.Encode(data); err != nil {
		return
	}
	encoded = buf.Bytes()
	return
}

func init() {
	mh.EncodeOptions = codec.EncodeOptions{Canonical: true}
}
