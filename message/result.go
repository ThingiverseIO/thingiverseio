package message

import (
	"bytes"

	"github.com/ThingiverseIO/uuid"
	"github.com/ugorji/go/codec"
)

//go:generate evt2gogen -t *Result -n Result

type Result struct {
	Request *Request
	Output  uuid.UUID
	params  []byte
}

func NewResult(output uuid.UUID, request *Request, parameter []byte) (r *Result) {
	r = &Result{
		Output:  output,
		Request: request,
		params:  parameter,
	}
	return
}

func (Result) GetType() Type { return RESULT }

func (Result) New() Message {
	return new(Result)
}

func (r *Result) Unflatten(d [][]byte) {
	dec := codec.NewDecoder(bytes.NewBuffer(d[0]), &mh)
	dec.Decode(&r)
	r.params = d[1]
}

func (r Result) Flatten() [][]byte {
	var payload bytes.Buffer
	enc := codec.NewEncoder(&payload, &mh)
	enc.Encode(r)
	return [][]byte{payload.Bytes(), r.params}
}

func (r Result) Parameter() []byte {
	return r.params
}

func (r Result) Decode(t interface{}) (err error) {
	buf := bytes.NewBuffer(r.params)
	dec := codec.NewDecoder(buf, &mh)
	err = dec.Decode(t)
	return
}

func init() {
	registerMessage(new(Result))
}
