package message

import "github.com/ugorji/go/codec"

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.EncodeOptions = codec.EncodeOptions{Canonical: true}
}
