//	along with libaursir.  If not, see <http://www.gnu.org/licenses/>.

package main

import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/joernweissenborn/thingiverseio"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

var nextInput = 0

var inputs = map[int]*thingiverseio.Input{}
var inputsLock = &sync.RWMutex{}

var request = map[int]map[config.UUID]*messages.ResultFuture{}
var requestLock = &sync.RWMutex{}

var listenResults = map[int]*messages.ResultCollector{}
var listenResultsLock = &sync.RWMutex{}

var callAllResults = map[int]map[config.UUID]*messages.ResultCollector{}
var callAllResultsLock = &sync.RWMutex{}

func getNextInput() (n int) {
	n = nextInput
	nextInput++
	return
}

func newInput(desc string) (n int) {
	inputsLock.Lock()
	defer inputsLock.Unlock()
	requestLock.Lock()
	defer requestLock.Unlock()
	listenResultsLock.Lock()
	defer listenResultsLock.Unlock()
	n = getNextInput()
	i, err := thingiverseio.NewInput(desc)
	if err != nil {
		return -1
	}
	inputs[n] = i
	request[n] = map[config.UUID]*messages.ResultFuture{}
	listenResults[n] = messages.NewResultCollector()
	listenResults[n].AddStream(i.ListenResults())
	i.Run()
	return
}

//export new_input
func new_input(descriptor *C.char) (i C.int) {
	d := C.GoString(descriptor)
	return C.int(newInput(d))
}

//export remove_input
func remove_input(i C.int) C.int {
	inputsLock.Lock()
	defer inputsLock.Unlock()
	requestLock.Lock()
	defer requestLock.Unlock()
	listenResultsLock.Lock()
	defer listenResultsLock.Unlock()
	callAllResultsLock.Lock()
	defer callAllResultsLock.Unlock()
	if in, ok := inputs[int(i)]; ok {
		in.Remove()
		delete(request, int(i))
		delete(listenResults, int(i))
		delete(callAllResults, int(i))
		delete(inputs, int(i))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export connected
func connected(i C.int, is *C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		if in.HasConnection() {
			*is = 1
		} else {
			*is = 0
		}
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export get_input_uuid
func get_input_uuid(i C.int, uuid **C.char, uuid_size *C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		*uuid = C.CString(string(in.UUID()))
		*uuid_size = C.int(len(in.UUID()))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export get_input_interface
func get_input_interface(i C.int, iface **C.char, iface_size *C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		*iface = C.CString(string(in.Interface()))
		*iface_size = C.int(len(in.Interface()))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export start_listen
func start_listen(i C.int, function *C.char) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		in.Listen(C.GoString(function))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export stop_listen
func stop_listen(i C.int, function *C.char) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		in.StopListen(C.GoString(function))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export call
func call(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int, request_id **C.char, request_id_size *C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	requestLock.Lock()
	defer requestLock.Unlock()
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := getParams(parameter, parameter_size)

		uuid, f := in.CallBin(fun, params)

		request[int(i)][uuid] = f

		*request_id = C.CString(string(uuid))
		*request_id_size = C.int(len(uuid))
		fmt.Println("blalawf", *request_id, *request_id_size)
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export call_all
func call_all(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int, request_id **C.char, request_id_size *C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	callAllResultsLock.Lock()
	defer callAllResultsLock.Unlock()
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := getParams(parameter, parameter_size)

		s := messages.NewResultStreamController()
		c := messages.NewResultCollector()
		c.AddStream(s.Stream())

		uuid := in.CallAllBin(fun, params, s)
		callAllResults[int(i)][uuid] = c
		*request_id = C.CString(string(uuid))
		*request_id_size = C.int(len(uuid))
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export trigger
func trigger(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := getParams(parameter, parameter_size)
		in.TriggerBin(fun, params)
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export trigger_all
func trigger_all(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	inputsLock.RLock()
	defer inputsLock.RUnlock()
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := getParams(parameter, parameter_size)
		in.TriggerAllBin(fun, params)
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export result_ready
func result_ready(i C.int, uuid *C.char, ready *C.int) C.int {
	requestLock.RLock()
	defer requestLock.RUnlock()
	if request[int(i)] == nil {
		return ERR_INVALID_INPUT
	}
	if f, ok := request[int(i)][config.UUID(C.GoString(uuid))]; ok {
		if f.Completed() {
			*ready = 1
		} else {
			*ready = 0
		}
		return C.int(0)
	}
	return ERR_INVALID_RESULT_ID
}

//export retrieve_result_params
func retrieve_result_params(i C.int, uuid *C.char, result *unsafe.Pointer, result_size *C.int) C.int {
	requestLock.Lock()
	defer requestLock.Unlock()
	if request[int(i)] == nil {
		return ERR_INVALID_INPUT
	}
	if f, ok := request[int(i)][config.UUID(C.GoString(uuid))]; ok {
		if !f.Completed() {
			return ERR_RESULT_NOT_ARRIVED
		}
		*result = unsafe.Pointer(C.CString(string(f.GetResult().Parameter())))
		*result_size = C.int(len(f.GetResult().Parameter()))
		delete(request[int(i)], config.UUID(C.GoString(uuid)))
		return C.int(0)
	}
	return ERR_INVALID_RESULT_ID
}

//export listen_result_available
func listen_result_available(i C.int, is *C.int) C.int {
	listenResultsLock.RLock()
	defer listenResultsLock.RUnlock()
	if res, ok := listenResults[int(i)]; ok {
		if res.Empty() {
			*is = 0
		} else {
			*is = 1
		}
		return C.int(0)
	}
	return ERR_INVALID_INPUT
}

//export retrieve_listen_result_id
func retrieve_listen_result_id(i C.int, request_id **C.char, request_id_size *C.int) C.int {
	listenResultsLock.RLock()
	defer listenResultsLock.RUnlock()
	if res, ok := listenResults[int(i)]; ok {
		if res.Empty() {
			return ERR_NO_RESULT_AVAILABLE
		} else {
			uuid := res.Preview().Request.UUID
			*request_id = C.CString(string(uuid))
			*request_id_size = C.int(len(uuid))
			return C.int(0)
		}
	}
	return ERR_INVALID_INPUT
}

//export retrieve_listen_result_function
func retrieve_listen_result_function(i C.int, function **C.char, function_size *C.int) C.int {
	listenResultsLock.RLock()
	defer listenResultsLock.RUnlock()
	if res, ok := listenResults[int(i)]; ok {
		if res.Empty() {
			return ERR_NO_RESULT_AVAILABLE
		} else {
			fun := res.Preview().Request.Function
			*function = C.CString(fun)
			*function_size = C.int(len(fun))
			return C.int(0)
		}
	}
	return ERR_INVALID_INPUT
}

//Nexport retrieve_listen_result_request_params
//TODO: rework scheme of getting request parameter
func retrieve_listen_result_request_params(i C.int, params *unsafe.Pointer, params_size *C.int) C.int {
	listenResultsLock.RLock()
	defer listenResultsLock.RUnlock()
	if res, ok := listenResults[int(i)]; ok {
		if res.Empty() {
			return ERR_NO_RESULT_AVAILABLE
		} else {
			p := res.Preview().Request.Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return ERR_INVALID_INPUT
}

//export retrieve_listen_result_params
func retrieve_listen_result_params(i C.int, params *unsafe.Pointer, params_size *C.int) C.int {
	listenResultsLock.RLock()
	defer listenResultsLock.RUnlock()
	if res, ok := listenResults[int(i)]; ok {
		if res.Empty() {
			return ERR_NO_RESULT_AVAILABLE
		} else {
			p := res.Get().Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return ERR_INVALID_INPUT
}

func call_all_result_available(i C.int, uuid *C.char, is *C.int) C.int {
	callAllResultsLock.RLock()
	defer callAllResultsLock.RUnlock()
	if r, ok := callAllResults[int(i)]; ok {
		if res, ok := r[config.UUID(C.GoString(uuid))]; ok {
			if res.Empty() {
				*is = C.int(0)
			} else {
				*is = C.int(1)
			}
		}
	}
	return ERR_INVALID_INPUT
}

//export retrieve_next_call_all_result_params
func retrieve_next_call_all_result_params(i C.int, uuid *C.char, params *unsafe.Pointer, params_size *C.int) C.int {
	callAllResultsLock.RLock()
	defer callAllResultsLock.RUnlock()
	if r, ok := callAllResults[int(i)]; ok {
		if res, ok := r[config.UUID(C.GoString(uuid))]; ok {
			if res.Empty() {
				return ERR_NO_RESULT_AVAILABLE
			}
			p := res.Get().Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return ERR_INVALID_INPUT
}
