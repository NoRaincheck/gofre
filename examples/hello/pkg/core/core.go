package core

//export Fibonacci
func Fibonacci(n int64) int64 {
	if n <= 1 {
		return n
	}
	return Fibonacci(n-1) + Fibonacci(n-2)
}

//export SumSlice
func SumSlice(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum
}

//export Multiply
func Multiply(a, b int64) int64 {
	return a * b
}

//export Add
func Add(a, b int64) int64 {
	return a + b
}
