package math

//export SumSlice
func SumSlice(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum
}

//export DoubleSlice
func DoubleSlice(data []int64) []int64 {
	result := make([]int64, len(data))
	for i, v := range data {
		result[i] = v * 2
	}
	return result
}
