//	Copyright (c) 2015 Joern Weissenborn
//
//	This file is part of libaursir.
//
//	Foobar is free software: you can redistribute it and/or modify
//	it under the terms of the GNU General Public License as published by
//	the Free Software Foundation, either version 3 of the License, or
//	(at your option) any later version.
//
//	libaursir is distributed in the hope that it will be useful,
//	but WITHOUT ANY WARRANTY; without even the implied warranty of
//	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//	GNU General Public License for more details.
//
//	You should have received a copy of the GNU General Public License
//	along with libaursir.  If not, see <http://www.gnu.org/licenses/>.

package main

import "C"

import (
	"unsafe"

	"github.com/joernweissenborn/thingiverseio"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

var nextOutput = 0

func getNextOutput() (n int) {
	n = nextOutput
	nextOutput++
	return
}

var outputs = map[int]*thingiverseio.Output{}

func newOutput(desc string) (n int) {
	n = getNextOutput()
	e, err := thingiverseio.NewOutput(desc)
	if err != nil {
		return -1
	}
	outputs[n] = e
	requestIn[n] = map[config.UUID]*messages.Request{}
	waiting_requestIn[n] = []config.UUID{}
	e.Requests().Listen(getRequest(n))
	e.Run()
	return
}

var requestIn = map[int]map[config.UUID]*messages.Request{}
var waiting_requestIn = map[int][]config.UUID{}

func getRequest(n int) messages.RequestSubscriber {
	return func(r *messages.Request) {
		waiting_requestIn[n] = append(waiting_requestIn[n], r.UUID)
		requestIn[n][r.UUID] = r
	}
}

//export new_output
func new_output(descriptor *C.char) C.int {
	d := C.GoString(descriptor)
	return C.int(newOutput(d))
}

//export get_next_request_id
func get_next_request_id(o C.int, uuid **C.char, uuid_size *C.int) C.int {
	if waiting, ok := waiting_requestIn[int(o)]; ok {
		if len(waiting) != 0 {
			*uuid = C.CString(string(waiting[0]))
			*uuid_size = C.int(len(waiting[0]))
			waiting_requestIn[int(o)] = waiting_requestIn[int(o)][1:]
		}
		return C.int(0)
	}
	return C.int(1)
}

//export retrieve_request_function
func retrieve_request_function(o C.int, uuid *C.char, function **C.char, function_size *C.int) C.int {
	if r, ok := requestIn[int(o)][config.UUID(C.GoString(uuid))]; ok {
		*function = C.CString(r.Function)
		*function_size = C.int(len(r.Function))
		return C.int(0)
	}
	return C.int(1)
}

//export retrieve_request_params
func retrieve_request_params(o C.int, uuid *C.char, parameter *unsafe.Pointer, parameter_size *C.int) C.int {
	if r, ok := requestIn[int(o)][config.UUID(C.GoString(uuid))]; ok {
		*parameter = unsafe.Pointer(C.CString(string(r.Parameter())))
		*parameter_size = C.int(len(r.Parameter()))
		return C.int(0)
	}
	return C.int(1)
}

//export reply
func reply(o C.int, uuid *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	r := requestIn[int(o)][config.UUID(C.GoString(uuid))]
	out := outputs[int(o)]
	if r != nil && out != nil {
		params := []byte(C.GoStringN((*C.char)(parameter), parameter_size))
		out.ReplyEncoded(r, params)
		delete(requestIn[int(o)], config.UUID(C.GoString(uuid)))
		return C.int(0)
	}
	return C.int(1)
}

//export emit
func emit(o C.int, function *C.char, inparameter unsafe.Pointer, inparameter_size C.int, outparameter unsafe.Pointer, outparameter_size C.int) C.int {
	out := outputs[int(o)]
	if out != nil {
		out.EmitEncoded(
			C.GoString(function),
			[]byte(C.GoStringN((*C.char)(inparameter), inparameter_size)),
			[]byte(C.GoStringN((*C.char)(outparameter), outparameter_size)),
		)
		return C.int(0)
	}
	return C.int(1)
}

