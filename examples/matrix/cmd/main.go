package main

// #include <stdint.h>
// #include <stdlib.h>
import "C"
import "unsafe"

import (
	"github.com/grackin/examples/matrix/pkg/matrix"
)

func main() {}

//export MatrixMultiply
func MatrixMultiply(a *C.double, aLen C.int64_t, b *C.double, bLen C.int64_t, n C.int64_t) *C.double {
	goA := unsafe.Slice((*float64)(a), int(aLen))
	goB := unsafe.Slice((*float64)(b), int(bLen))
	result := matrix.MatrixMultiply(goA, goB, int64(n))
	size := int(n) * int(n)
	ptr := (*C.double)(C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(C.double(0)))))
	copy(unsafe.Slice((*float64)(ptr), size), result)
	return ptr
}

//export MatrixTranspose
func MatrixTranspose(m *C.double, mLen C.int64_t, rows C.int64_t, cols C.int64_t) *C.double {
	goM := unsafe.Slice((*float64)(m), int(mLen))
	result := matrix.MatrixTranspose(goM, int64(rows), int64(cols))
	size := int(rows) * int(cols)
	ptr := (*C.double)(C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(C.double(0)))))
	copy(unsafe.Slice((*float64)(ptr), size), result)
	return ptr
}

//export DotProduct
func DotProduct(a *C.double, aLen C.int64_t, b *C.double, bLen C.int64_t) C.double {
	goA := unsafe.Slice((*float64)(a), int(aLen))
	goB := unsafe.Slice((*float64)(b), int(bLen))
	return C.double(matrix.DotProduct(goA, goB))
}
