//	along with libaursir.  If not, see <http://www.gnu.org/licenses/>.

package main

import "C"

import (
	"github.com/joernweissenborn/eventual2go"
	"github.com/joernweissenborn/thingiverseio"
	"github.com/joernweissenborn/thingiverseio/service/messages"
)

var nextInput = 0

var request = map[int]map[string]*eventual2go.Future{}
var listen_results = map[int]*eventual2go.Collector{}
var n2n_streams = map[int]map[string]*eventual2go.Collector{}

func getNextInput() (n int) {
	n = nextInput
	nextInput++
	return
}

var inputs = map[int]*thingiverseio.Input{}

func newImport(desc *thingiverseio.Descriptor) (n int) {
	n = getNextInput()
	i, err := thingiverseio.NewInput(desc)
	if err != nil {
		return -1
	}
	inputs[n] = i
	request[n] = map[string]*eventual2go.Future{}
	listen_results[n] = eventual2go.NewCollector()
	listen_results[n].AddStream(i.Results())
	i.Run()
	return
}

//export NewInput
func NewInput(desc *C.char) (i C.int) {
	return C.int(newImport(desc))
}

//export Listen
func Listen(i C.int, function *C.char) {
	if in, ok := inputs[int(i)]; ok {
		in.Listen(C.GoString(function))
	}
}

//export StopListen
func StopListen(i C.int, function *C.char) {
	if in, ok := inputs[int(i)]; ok {
		in.StopListen(C.GoString(function))
	}
}

//export Call
func Call(i C.int, function *C.char, parameter *C.char) *C.char {
	in := inputs[int(i)]
	if in != nil {
		req := in.NewRequestBin(C.GoString(function), []byte(C.GoString(parameter)), messages.ONE2ONE)

		isRes := func(uuid string) eventual2go.Filter {
			return func(d eventual2go.Data) bool {
				return d.(*messages.Result).Request.UUID == uuid
			}
		}

		request[int(i)][req.UUID] = in.Results().FirstWhere(isRes(req.UUID))
		in.Deliver(req)
		return C.CString(req.UUID)
	}
	return nil
}

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

//export GetResult
func GetResult(i C.int, uuid *C.char) *C.char {
	if request[int(i)] == nil {
		return nil
	} else if len(request[int(i)]) == 0 {
		return nil
	} else {
		f := request[int(i)][C.GoString(uuid)]
		if f == nil {
			return nil
		}
		if !f.Completed() {
			return nil
		}
		return C.CString(string(f.GetResult().(*messages.Result).Parameter()))
	}
}

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
