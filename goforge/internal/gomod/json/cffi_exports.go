//go:build cffi

package jsonbridge

/*
#include <stdlib.h>
*/
import "C"
import "unsafe"

//export JSONDumps
func JSONDumps(obj *C.char) *C.char {
	goObj := C.GoString(obj)
	result := GoDumps(goObj)
	return C.CString(result)
}

//export JSONLoads
func JSONLoads(s *C.char) *C.char {
	goS := C.GoString(s)
	result := GoLoads(goS)
	return C.CString(result)
}

func init() {
	_ = unsafe.Pointer(nil)
}
