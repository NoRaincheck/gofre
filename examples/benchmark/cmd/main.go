package main

// #include <stdint.h>
// #include <stdlib.h>
import "C"
import "unsafe"

import (
	benchmarkmath "github.com/grackin/examples/benchmark/pkg/math"
)

func main() {}

//export Fibonacci
func Fibonacci(n C.int64_t) C.int64_t {
	return C.int64_t(benchmarkmath.Fibonacci(int64(n)))
}

//export SumSlice
func SumSlice(data *C.double, dataLen C.int64_t) C.double {
	goData := unsafe.Slice((*float64)(data), int(dataLen))
	return C.double(benchmarkmath.SumSlice(goData))
}

//export MatrixMultiply
func MatrixMultiply(a *C.double, aLen C.int64_t, b *C.double, bLen C.int64_t, n C.int64_t) *C.double {
	goA := unsafe.Slice((*float64)(a), int(aLen))
	goB := unsafe.Slice((*float64)(b), int(bLen))
	result := benchmarkmath.MatrixMultiply(goA, goB, int64(n))
	size := int(n) * int(n)
	ptr := (*C.double)(C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(C.double(0)))))
	copy(unsafe.Slice((*float64)(ptr), size), result)
	return ptr
}

//export CountPrimes
func CountPrimes(limit C.int64_t) C.int64_t {
	return C.int64_t(benchmarkmath.CountPrimes(int64(limit)))
}

//export SortInts
func SortInts(data *C.int64_t, dataLen C.int64_t) *C.int64_t {
	goData := unsafe.Slice((*int64)(data), int(dataLen))
	result := benchmarkmath.SortInts(goData)
	size := len(result)
	ptr := (*C.int64_t)(C.malloc(C.size_t(size) * C.size_t(unsafe.Sizeof(C.int64_t(0)))))
	copy(unsafe.Slice((*int64)(ptr), size), result)
	return ptr
}
