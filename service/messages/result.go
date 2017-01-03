package messages

import (
	"bytes"

	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ugorji/go/codec"
)

//go:generate event_generator -t *Result -n Result

type Result struct {
	Request *Request
	Output  config.UUID
	params  []byte
}

func NewResult(output config.UUID, request *Request, parameter interface{}) (r *Result) {
	var params bytes.Buffer
	enc := codec.NewEncoder(&params, &mh)
	enc.Encode(parameter)
	return NewEncodedResult(output, request, params.Bytes())
}

func NewEncodedResult(output config.UUID, request *Request, parameter []byte) (r *Result) {
	r = new(Result)
	r.Output = output
	r.Request = request
	r.params = parameter
	return
}

func (*Result) GetType() MessageType { return RESULT }

func (*Result) New() Message {
	return new(Result)
}

func (r *Result) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(r)
	r.params = d[1]
}

func (r *Result) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(r)
	return [][]byte{payload.Bytes(), r.params}
}

func (r *Result) Parameter() []byte {
	return r.params
}

func (r *Result) Decode(t interface{}) {
	buf := bytes.NewBuffer(r.params)
	dec := codec.NewDecoder(buf, &mh)
	dec.Decode(t)
	return
}

func init() {
	registerMessage(new(Result))
}
