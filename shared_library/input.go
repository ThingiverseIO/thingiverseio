//	along with libaursir.  If not, see <http://www.gnu.org/licenses/>.

package main

import "C"

import (
	"unsafe"

	"github.com/joernweissenborn/thingiverseio"
	"github.com/joernweissenborn/thingiverseio/config"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

var nextInput = 0

var request = map[int]map[config.UUID]*messages.ResultFuture{}
var listen_results = map[int]*messages.ResultCollector{}
var call_all_results = map[int]map[config.UUID]*messages.ResultCollector{}

func getNextInput() (n int) {
	n = nextInput
	nextInput++
	return
}

var inputs = map[int]*thingiverseio.Input{}

func newInput(desc string) (n int) {
	n = getNextInput()
	i, err := thingiverseio.NewInput(desc)
	if err != nil {
		return -1
	}
	inputs[n] = i
	request[n] = map[config.UUID]*messages.ResultFuture{}
	listen_results[n] = messages.NewResultCollector()
	listen_results[n].AddStream(i.ListenResults())
	i.Run()
	return
}

//export new_input
func new_input(descriptor *C.char) (i C.int) {
	d := C.GoString(descriptor)
	return C.int(newInput(d))
}

//export start_listen
func start_listen(i C.int, function *C.char) C.int {
	if in, ok := inputs[int(i)]; ok {
		in.Listen(C.GoString(function))
		return C.int(0)
	}
	return C.int(1)
}

//export stop_listen
func stop_listen(i C.int, function *C.char) C.int {
	if in, ok := inputs[int(i)]; ok {
		in.StopListen(C.GoString(function))
		return C.int(0)
	}
	return C.int(1)
}

//export call
func call(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int, request_id **C.char, request_id_size *C.int) C.int {
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := []byte(C.GoStringN((*C.char)(parameter), parameter_size))
		uuid, f := in.CallBin(fun, params)

		request[int(i)][uuid] = f

		*request_id = C.CString(string(uuid))
		*request_id_size = C.int(len(uuid))
		return C.int(0)
	}
	return C.int(1)
}

//export call_all
func call_all(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int, request_id **C.char, request_id_size *C.int) C.int {
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := []byte(C.GoStringN((*C.char)(parameter), parameter_size))

		s := messages.NewResultStreamController()
		c := messages.NewResultCollector()
		c.AddStream(s.Stream())

		uuid := in.CallAllBin(fun, params, s)
		call_all_results[int(i)][uuid] = c
		*request_id = C.CString(string(uuid))
		*request_id_size = C.int(len(uuid))
		return C.int(0)
	}
	return C.int(1)
}

//export trigger
func trigger(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := []byte(C.GoStringN((*C.char)(parameter), parameter_size))
		in.TriggerBin(fun, params)
		return C.int(0)
	}
	return C.int(1)
}

//export trigger_all
func trigger_all(i C.int, function *C.char, parameter unsafe.Pointer, parameter_size C.int) C.int {
	if in, ok := inputs[int(i)]; ok {
		fun := C.GoString(function)
		params := []byte(C.GoStringN((*C.char)(parameter), parameter_size))
		in.TriggerAllBin(fun, params)
		return C.int(0)
	}
	return C.int(1)
}

//export result_ready
func result_ready(i C.int, uuid *C.char, ready *C.int) C.int {
	if request[int(i)] == nil {
		return C.int(1)
	}
	if f, ok := request[int(i)][config.UUID(C.GoString(uuid))]; ok {
		if f.Completed() {
			*ready = 1
		} else {
			*ready = 0
		}
		return C.int(0)
	}
	return C.int(1)
}

//export retrieve_result_params
func retrieve_result_params(i C.int, uuid *C.char, result *unsafe.Pointer, result_size *C.int) C.int {

	if request[int(i)] == nil {
		return C.int(1)
	}
	if f, ok := request[int(i)][config.UUID(C.GoString(uuid))]; ok {
		if !f.Completed() {
			return C.int(1)
		}
		*result = unsafe.Pointer(C.CString(string(f.GetResult().Parameter())))
		*result_size = C.int(len(f.GetResult().Parameter()))
		return C.int(0)
	}
	return C.int(1)
}

//export listen_result_available
func listen_result_available(i C.int, is *C.int) C.int {
	if res, ok := listen_results[int(i)]; ok {
		if res.Empty() {
			*is = 0
		} else {
			*is = 1
		}
		return C.int(0)
	}
	return C.int(1)
}

//export retrieve_listen_result_id
func retrieve_listen_result_id(i C.int, request_id **C.char, request_id_size *C.int) C.int {
	if res, ok := listen_results[int(i)]; ok {
		if res.Empty() {
			return C.int(1)
		} else {
			uuid := res.Preview().Request.UUID
			*request_id = C.CString(string(uuid))
			*request_id_size = C.int(len(uuid))
			return C.int(0)
		}
	}
	return C.int(1)
}

//export retrieve_listen_result_function
func retrieve_listen_result_function(i C.int, function **C.char, function_size *C.int) C.int {
	if res, ok := listen_results[int(i)]; ok {
		if res.Empty() {
			return C.int(1)
		} else {
			fun := res.Preview().Request.Function
			*function = C.CString(fun)
			*function_size = C.int(len(fun))
			return C.int(0)
		}
	}
	return C.int(1)
}

//export retrieve_listen_result_request_params
func retrieve_listen_result_request_params(i C.int, params *unsafe.Pointer, params_size *C.int) C.int {
	if res, ok := listen_results[int(i)]; ok {
		if res.Empty() {
			return C.int(1)
		} else {
			p := res.Preview().Request.Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return C.int(1)
}

//export retrieve_listen_result_params
func retrieve_listen_result_params(i C.int, params *unsafe.Pointer, params_size *C.int) C.int {
	if res, ok := listen_results[int(i)]; ok {
		if res.Empty() {
			return C.int(1)
		} else {
			p := res.Get().Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return C.int(1)
}

func call_all_result_available(i C.int, uuid *C.char, is *C.int) C.int {
	if r, ok := call_all_results[int(i)]; ok {
		if res, ok := r[config.UUID(C.GoString(uuid))]; ok {
			if res.Empty() {
				*is = C.int(0)
			} else {
				*is = C.int(1)
			}
		}
	}
	return C.int(1)
}

//export retrieve_next_call_all_result_params
func retrieve_next_call_all_result_params(i C.int, uuid *C.char, params *unsafe.Pointer, params_size *C.int) C.int {
	if r, ok := call_all_results[int(i)]; ok {
		if res, ok := r[config.UUID(C.GoString(uuid))]; ok {
			p := res.Get().Parameter()
			*params = unsafe.Pointer(C.CString(string(p)))
			*params_size = C.int(len(p))
			return C.int(0)
		}
	}
	return C.int(1)
}
