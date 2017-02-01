package message

import (
	"bytes"

	"github.com/ugorji/go/codec"
)

type StartListen struct {
	Function string
}

func (*StartListen) New() Message {
	return new(StartListen)
}

func (*StartListen) GetType() Type { return STARTLISTEN }

func (l *StartListen) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(l)
}

func (l *StartListen) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(l)
	return [][]byte{payload.Bytes()}
}

func init() {
	registerMessage(new(StartListen))
}
