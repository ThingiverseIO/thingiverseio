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
	"sync"
	"unsafe"

	"github.com/ThingiverseIO/thingiverseio"
	"github.com/ThingiverseIO/thingiverseio/config"
	"github.com/ThingiverseIO/thingiverseio/service/messages"
)

var nextOutput = 0

func getNextOutput() (n int) {
	n = nextOutput
	nextOutput++
	return
}

var outputs = map[int]*thingiverseio.Output{}
var outputLock = &sync.RWMutex{}

var requestIn = map[int]map[config.UUID]*messages.Request{}
var requestInLock = &sync.RWMutex{}

var waitingRequests = map[int]*config.UUIDCollector{}
var waitingRequestsLock = &sync.RWMutex{}

func newOutput(desc string) (n int) {
	outputLock.Lock()
	defer outputLock.Unlock()
	requestInLock.Lock()
	defer requestInLock.Unlock()
	waitingRequestsLock.Lock()
	defer waitingRequestsLock.Unlock()
	n = getNextOutput()
	e, err := thingiverseio.NewOutput(desc)
	if err != nil {
		return -1
	}
	outputs[n] = e
	requestIn[n] = map[config.UUID]*messages.Request{}
	waitingRequests[n] = config.NewUUIDCollector()
	e.Requests().Listen(getRequest(n))
	e.Run()
	return
}

func getRequest(n int) messages.RequestSubscriber {
	return func(r *messages.Request) {
		requestInLock.Lock()
		defer requestInLock.Unlock()
		waitingRequestsLock.Lock()
		defer waitingRequestsLock.Unlock()
		waitingRequests[n].Add(r.UUID)
		requestIn[n][r.UUID] = r
	}
}

//export new_output
func new_output(descriptor *C.char) C.int {
	d := C.GoString(descriptor)
	return C.int(newOutput(d))
}

//export remove_output
func remove_output(o C.int) C.int {
	outputLock.Lock()
	defer outputLock.Unlock()
	requestInLock.Lock()
	defer requestInLock.Unlock()
	waitingRequestsLock.Lock()
	defer waitingRequestsLock.Unlock()
	if out, ok := outputs[int(o)]; ok {
		out.Remove()
		delete(outputs, int(o))
		delete(requestIn, int(o))
		waitingRequests[int(o)].Remove()
		delete(waitingRequests, int(o))
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export get_output_uuid
func get_output_uuid(o C.int, uuid **C.char, uuid_size *C.int) C.int {
	outputLock.RLock()
	defer outputLock.RUnlock()
	if out, ok := inputs[int(o)]; ok {
		*uuid = C.CString(string(out.UUID()))
		*uuid_size = C.int(len(out.UUID()))
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export get_output_interface
func get_output_interface(o C.int, iface **C.char, iface_size *C.int) C.int {
	outputLock.RLock()
	defer outputLock.RUnlock()
	if out, ok := inputs[int(o)]; ok {
		*iface = C.CString(string(out.UUID()))
		*iface_size = C.int(len(out.UUID()))
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export get_next_request_id
func get_next_request_id(o C.int, uuid **C.char, uuid_size *C.int) C.int {
	waitingRequestsLock.RLock()
	defer waitingRequestsLock.RUnlock()
	if waiting, ok := waitingRequests[int(o)]; ok {
		if !waiting.Empty() {
			*uuid = C.CString(string(waiting.Preview()))
			*uuid_size = C.int(len(waiting.Get()))
		}
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export request_available
func request_available(o C.int, is *C.int) C.int {
	waitingRequestsLock.RLock()
	defer waitingRequestsLock.RUnlock()
	if waiting, ok := waitingRequests[int(o)]; ok {
		if !waiting.Empty() {
			*is = C.int(1)
		} else {
			*is = C.int(0)
		}
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export retrieve_request_function
func retrieve_request_function(o C.int, uuid *C.char, function **C.char, function_size *C.int) C.int {
	requestInLock.RLock()
	defer requestInLock.RUnlock()
	if r, ok := requestIn[int(o)][config.UUID(C.GoString(uuid))]; ok {
		*function = C.CString(r.Function)
		*function_size = C.int(len(r.Function))
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export retrieve_request_params
func retrieve_request_params(o C.int, uuid *C.char, parameter *unsafe.Pointer, parameter_size *C.int) C.int {
	requestInLock.RLock()
	defer requestInLock.RUnlock()
	if r, ok := requestIn[int(o)][config.UUID(C.GoString(uuid))]; ok {
		*parameter = unsafe.Pointer(C.CString(string(r.Parameter())))
		*parameter_size = C.int(len(r.Parameter()))
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}

//export reply
func reply(o C.int, uuid *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	outputLock.RLock()
	defer outputLock.RUnlock()
	requestInLock.Lock()
	defer requestInLock.Unlock()
	out := outputs[int(o)]
	if out == nil {
		return ERR_INVALID_OUTPUT
	}
	r := requestIn[int(o)][config.UUID(C.GoString(uuid))]
	if r != nil {
		params := getParams(parameter, parameter_size)
		out.ReplyEncoded(r, params)
		delete(requestIn[int(o)], config.UUID(C.GoString(uuid)))
		return C.int(0)
	}
	return ERR_INVALID_REQUEST_ID
}

//export emit
func emit(o C.int, function *C.char, inparameter unsafe.Pointer, inparameter_size C.int, outparameter unsafe.Pointer, outparameter_size C.int) C.int {
	outputLock.RLock()
	defer outputLock.RUnlock()
	out := outputs[int(o)]
	if out != nil {
		out.EmitEncoded(
			C.GoString(function),
			getParams(inparameter, inparameter_size),
			getParams(outparameter, outparameter_size),
		)
		return C.int(0)
	}
	return ERR_INVALID_OUTPUT
}
