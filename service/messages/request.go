package messages

import (
	"bytes"
	"strings"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t *Request -n Request

type Request struct {
	UUID     config.UUID
	Input    config.UUID
	Function string
	CallType CallType
	params   []byte
}

func NewRequest(input config.UUID, function string, call_type CallType, parameter interface{}) (r *Request) {
	var params bytes.Buffer
	enc := codec.NewEncoder(&params, &mh)
	enc.Encode(parameter)
	return NewEncodedRequest(input, function, call_type, params.Bytes())
}

func NewEncodedRequest(input config.UUID, function string, call_type CallType, params []byte) (r *Request) {
	return NewEncodedRequestWithId(config.NewUUID(), input, function, call_type, params)
}

func NewEncodedRequestWithId(uuid, input config.UUID, function string, call_type CallType, params []byte) (r *Request) {
	r = &Request{
		UUID:     uuid,
		Input:    input,
		Function: function,
		CallType: call_type,
		params:   params,
	}
	return
}

func (*Request) GetType() MessageType { return REQUEST }

func (*Request) New() Message {
	return new(Request)
}

func (r *Request) Unflatten(d []string) {
	dec := codec.NewDecoder(strings.NewReader(d[0]), &mh)
	dec.Decode(r)
	r.params = []byte(d[1])
}

func (r *Request) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(r)
	return [][]byte{payload.Bytes(), r.params}
}

func (r *Request) Parameter() []byte {
	return r.params
}

func (r *Request) Decode(t interface{}) {
	buf := bytes.NewBuffer(r.params)
	dec := codec.NewDecoder(buf, &mh)
	dec.Decode(t)

	return
}

func init() {
	registerMessage(new(Request))
}
