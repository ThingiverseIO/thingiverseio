package messages

import (
	"bytes"
	"strings"

	"github.com/ugorji/go/codec"
)

type StopListen struct {
	Function string
}

func (*StopListen) GetType() MessageType { return STOP_LISTEN }

func (l *StopListen) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(l)
}

func (l *StopListen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}
