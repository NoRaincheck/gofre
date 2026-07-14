package math

//export Add
func Add(a, b int64) int64 {
	return a + b
}

//export Multiply
func Multiply(a, b int64) int64 {
	return a * b
}

//export GetHello
func GetHello() string {
	return "hello"
}
