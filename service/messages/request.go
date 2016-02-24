package messages

import (
	"bytes"
	"strings"

	"github.com/joernweissenborn/thingiverse.io/config"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t *Request -n Result

type Request struct {
	UUID     config.UUID
	Importer string
	Function string
	CallType CallType
	params   []byte
}

func NewRequest(importer, function string, call_type CallType, parameter interface{}) (r *Request) {
	var params bytes.Buffer
	enc := codec.NewEncoder(&params, &mh)
	enc.Encode(parameter)
	return NewEncodedRequest(importer, function, call_type, params.Bytes())
}

func NewEncodedRequest(importer, function string, call_type CallType, params []byte) (r *Request) {
	r = new(Request)
	return NewEncodedRequestWithId(config.NewUUID(), importer, function, call_type, params)
}

func NewEncodedRequestWithId(uuid config.UUID, importer, function string, call_type CallType, params []byte) (r *Request) {
	r = &Request{
		UUID:     uuid,
		Importer: importer,
		Function: function,
		CallType: call_type,
		params:   params,
	}
	return
}

func (*Request) GetType() MessageType { return REQUEST }

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
