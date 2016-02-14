package messages

import (
	"strconv"

	"github.com/joernweissenborn/thingiverse.io/service"
	"github.com/ugorji/go/codec"
)

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.EncodeOptions = codec.EncodeOptions{Canonical: true}
}

type Message interface {
	GetType() MessageType
	Flatten() [][]byte
	Unflatten([]string)
}

func Flatten(m Message) [][]byte {
	t := strconv.FormatInt(int64(m.GetType()), 10)
	payload := [][]byte{[]byte{byte(service.PROTOCOLL_SIGNATURE)}, []byte(t)}
	for _, p := range m.Flatten() {
		payload = append(payload, p)
	}
	return payload
}

func Unflatten(m []string) (msg Message) {
	mtype, _ := strconv.ParseInt(m[1], 10, 8)
	msg = Get(MessageType(mtype))
	msg.Unflatten(m[2:])
	return
}
