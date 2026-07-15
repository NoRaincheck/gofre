//go:build cffi

package httpbridge

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/NoRaincheck/gofre/internal/gomod/json"
)

//export JSONDumps
func JSONDumps(obj *C.char) *C.char {
	goObj := C.GoString(obj)
	result := jsonbridge.GoDumps(goObj)
	return C.CString(result)
}

//export JSONLoads
func JSONLoads(s *C.char) *C.char {
	goS := C.GoString(s)
	result := jsonbridge.GoLoads(goS)
	return C.CString(result)
}

//export HTTPCreateServer
func HTTPCreateServer() C.int {
	return C.int(NewServerHandle())
}

//export HTTPAddRoute
func HTTPAddRoute(serverID C.int, method *C.char, path *C.char, handlerID C.int) {
	m := C.GoString(method)
	p := C.GoString(path)
	s := GetServer(int(serverID))
	if s != nil {
		s.Handle(m, p, nil)
	}
}

func init() {
	_ = unsafe.Pointer(nil)
}
