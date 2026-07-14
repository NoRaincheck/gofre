package main

// #include <stdint.h>
// #include <stdlib.h>
import "C"
import "unsafe"

import (
	"github.com/grackin/examples/hello/pkg/core"
)

func main() {}

//export Fibonacci
func Fibonacci(n C.int64_t) C.int64_t {
	return C.int64_t(core.Fibonacci(int64(n)))
}

//export SumSlice
func SumSlice(data *C.double, dataLen C.int64_t) C.double {
	goData := unsafe.Slice((*float64)(data), int(dataLen))
	return C.double(core.SumSlice(goData))
}

//export Multiply
func Multiply(a C.int64_t, b C.int64_t) C.int64_t {
	return C.int64_t(core.Multiply(int64(a), int64(b)))
}

//export Add
func Add(a C.int64_t, b C.int64_t) C.int64_t {
	return C.int64_t(core.Add(int64(a), int64(b)))
}
