package messages

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StopListen struct {
	Function string
}

func (*StopListen) New() Message {
	return new(StopListen)
}
func (*StopListen) GetType() MessageType { return STOPLISTEN }

func (l *StopListen) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(l)
}

func (l *StopListen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StopListen))
}
