package messages

import (
	"bytes"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ugorji/go/codec"
	"gopkg.in/vmihailenco/msgpack.v2"
	"strings"
)

type Request struct {
	UUID     string
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
	id, _ := uuid.NewV4()
	return NewEncodedRequestWithId(id.String(), importer, function, call_type, params)
}

func NewEncodedRequestWithId(uuid, importer, function string, call_type CallType, params []byte) (r *Request) {
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
	dec := msgpack.NewDecoder(strings.NewReader(d[0]))
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
	msgpack.Unmarshal(r.params, t)
	return
}
