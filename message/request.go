package message

import (
	"bytes"

	"github.com/ThingiverseIO/thingiverseio/uuid"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t *Request -n Request

type Request struct {
	UUID     uuid.UUID
	Input    uuid.UUID
	Function string
	CallType CallType
	params   []byte
}

func NewRequest(input uuid.UUID, function string, call_type CallType, parameter interface{}) (r Request) {
	var params bytes.Buffer
	enc := codec.NewEncoder(&params, &mh)
	enc.Encode(parameter)
	return NewEncodedRequest(input, function, call_type, params.Bytes())
}

func NewEncodedRequest(input uuid.UUID, function string, call_type CallType, params []byte) (r Request) {
	return NewEncodedRequestWithId(uuid.New(), input, function, call_type, params)
}

func NewEncodedRequestWithId(uuid, input uuid.UUID, function string, call_type CallType, params []byte) (r Request) {
	r = Request{
		UUID:     uuid,
		Input:    input,
		Function: function,
		CallType: call_type,
		params:   params,
	}
	return
}

func (Request) GetType() Type { return REQUEST }

func (Request) New() Message {
	return new(Request)
}

func (r Request) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(r)
	r.params = d[1]
}

func (r Request) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(r)
	return [][]byte{payload.Bytes(), r.params}
}

func (r Request) Parameter() []byte {
	return r.params
}

func (r Request) Decode(t interface{}) {
	buf := bytes.NewBuffer(r.params)
	dec := codec.NewDecoder(buf, &mh)
	dec.Decode(t)

	return
}

func init() {
	registerMessage(new(Request))
}
