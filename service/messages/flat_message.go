package messages

import (
	"bytes"

	"github.com/ThingiverseIO/thingiverseio/service"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t FlatMessage

type FlatMessage struct {
	Type    MessageType
	Payload [][]byte
}

func (f FlatMessage) Encode() []byte {
	var payload bytes.Buffer
	payload.Write([]byte{byte(service.PROTOCOLL_SIGNATURE)})
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(f)
	return payload.Bytes()
}

func Decode(payload []byte) (f FlatMessage, ok bool) {
	if ok = payload[0] == service.PROTOCOLL_SIGNATURE; !ok {
		return
	}
	buf := bytes.NewBuffer(payload[1:])
	dec := codec.NewDecoder(buf, &mh)
	dec.Decode(&f)
	return
}
