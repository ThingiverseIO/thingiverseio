package main

import "C"

import (
	"unsafe"

	"github.com/joernweissenborn/thingiverseio"
)

func main() {
}

//export version
func version(maj, min, fix *C.int) C.int {
	*maj = C.int(thingiverseio.CurrentVersion.Major)
	*min = C.int(thingiverseio.CurrentVersion.Minor)
	*fix = C.int(thingiverseio.CurrentVersion.Fix)
	return C.int(0)
}

func getParams(parameter unsafe.Pointer, parameter_size C.int) (params []byte) {
	params = make([]byte, parameter_size)
	if parameter_size > 0 {
		cparams := []byte(C.GoStringN((*C.char)(parameter), parameter_size))
		for i, b := range cparams {
			params[i] = b
		}
	}
	return
}
