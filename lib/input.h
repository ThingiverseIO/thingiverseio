/* Created by "go tool cgo" - DO NOT EDIT. */

/* package command-line-arguments */

/* Start of preamble from import "C" comments.  */





/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef __SIZE_TYPE__ GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

typedef struct { const char *p; GoInt n; } GoString;
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif


extern int new_input(char* p0);

extern int start_listen(int p0, char* p1);

extern int stop_listen(int p0, char* p1);

extern int call(int p0, char* p1, void* p2, int p3, char** p4, int* p5);

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

extern char* GetResult(int p0, char* p1);

extern int new_output(char* p0);

extern int get_next_requestIn_id(int p0, char** p1, int* p2);

extern int retrieve_requestIn_function(int p0, char* p1, char** p2, int* p3);

extern int retrieve_requestIn_params(int p0, char* p1, void** p2, int* p3);

extern int reply(int p0, char* p1, void* p2, int p3);

extern int emit(int p0, char* p1, void* p2, int p3, void* p4, int p5);

#ifdef __cplusplus
}
#endif
