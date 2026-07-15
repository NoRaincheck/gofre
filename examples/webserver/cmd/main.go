package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/grackin/examples/webserver/pkg/core"
)

func main() {}

//export Dumps
func Dumps(obj *C.char) *C.char {
	result := core.Dumps(C.GoString(obj))
	return C.CString(result)
}

//export Loads
func Loads(s *C.char) *C.char {
	result := core.Loads(C.GoString(s))
	return C.CString(result)
}

func init() {
	_ = unsafe.Pointer(nil)
}
