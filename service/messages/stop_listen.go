package messages

import (
	"bytes"
	"strings"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type StopListen struct {
	Function string
}

func (*StopListen) GetType() MessageType { return STOP_LISTEN }

func (l *StopListen) Unflatten(d []string) {
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
	dec.Decode(l)
}

func (l *StopListen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := msgpack.NewEncoder(&payload)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}
