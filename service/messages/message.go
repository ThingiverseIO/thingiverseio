package messages

import "github.com/ugorji/go/codec"

var (
	mh codec.MsgpackHandle
)

func init() {
	mh.EncodeOptions = codec.EncodeOptions{Canonical: true}
}

//go:generate event_generator -t Message

type Message interface {
	New() Message
	GetType() MessageType
	Flatten() [][]byte
	Unflatten([][]byte)
}

func Flatten(m Message) *FlatMessage {
	return &FlatMessage{
		Type:    m.GetType(),
		Payload: m.Flatten(),
	}
}

func Unflatten(m FlatMessage) (msg Message) {
	msg = Get(MessageType(m.Type))
	msg.Unflatten(m.Payload)
	return
}
