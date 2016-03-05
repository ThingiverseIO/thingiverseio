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
var n2n_streams = map[int]map[string]*messages.ResultCollector{}

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

/*
//export CallAll
func CallAll(i C.int, function *C.char, parameter *C.char) *C.char {
	in := inputs[int(i)]
	if in != nil {
		req := in.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.ONE2MANY)

		isRes := func(uuid string) eventual2go.Filter {
			return func(d eventual2go.Data) bool {
				return d.(*messages.Result).Request.UUID == uuid
			}
		}

		n2n_streams[int(i)][req.UUID] = eventual2go.NewCollector()
		n2n_streams[int(i)][req.UUID].AddStream(in.Results().Where(isRes(req.UUID)))
		in.Deliver(req)
		return C.CString(req.UUID)
	}
	return nil
}

//export Trigger
func Trigger(i C.int, function *C.char, parameter *C.char) {
	in := inputs[int(i)]
	if in != nil {
		in.Deliver(in.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.MANY2ONE))
	}
}

//export TriggerAll
func TriggerAll(i C.int, function *C.char, parameter *C.char) {
	in := inputs[int(i)]
	if in != nil {
		in.Deliver(in.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.MANY2MANY))
	}
}
*/

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

/*
//export GetNextListenResult
func GetNextListenResult(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := listen_results[int(i)].Preview().(*messages.Result).Request.UUID
		return C.CString(r)
	}
}

//export GetNextListenResultFunction
func GetNextListenResultFunction(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := listen_results[int(i)].Preview().(*messages.Result).Request.Function
		return C.CString(r)
	}
}

//export GetNextListenResultInParameter
func GetNextListenResultInParameter(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := string(listen_results[int(i)].Preview().(*messages.Result).Request.Parameter())
		return C.CString(r)
	}
}

//export GetNextListenResultParameter
func GetNextListenResultParameter(i C.int) *C.char {
	if listen_results[int(i)].Empty() {
		return nil
	} else {
		r := string(listen_results[int(i)].Get().(*messages.Result).Parameter())
		return C.CString(r)
	}
}

//export GetNextCallAllResultParameter
func GetNextCallAllResultParameter(i C.int, uuid *C.char) *C.char {
	if n2n_streams[int(i)] == nil {
		return nil
	} else if n2n_streams[int(i)][C.GoString(uuid)] == nil {
		return nil
	} else if n2n_streams[int(i)][C.GoString(uuid)].Empty() {
		return nil
	} else {
		r := string(n2n_streams[int(i)][C.GoString(uuid)].Get().(*messages.Result).Parameter())
		return C.CString(r)
	}
}
func main() {
	//	i := NewImportYAML(new(C.char),C.CString("127.0.0.1"))
	//	fmt.Println(C.GoString(Call(i,C.CString("SayHello"),C.CString(""))))
}
*/
