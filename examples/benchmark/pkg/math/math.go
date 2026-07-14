package math

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

//export MatrixMultiply
func MatrixMultiply(a, b []float64, n int64) []float64 {
	result := make([]float64, n*n)
	for i := int64(0); i < n; i++ {
		for j := int64(0); j < n; j++ {
			sum := 0.0
			for k := int64(0); k < n; k++ {
				sum += a[i*n+k] * b[k*n+j]
			}
			result[i*n+j] = sum
		}
	}
	return result
}

//export CountPrimes
func CountPrimes(limit int64) int64 {
	count := int64(0)
	for i := int64(2); i < limit; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

func isPrime(n int64) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := int64(3); i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

//export SortInts
func SortInts(data []int64) []int64 {
	result := make([]int64, len(data))
	copy(result, data)
	quickSort(result, 0, len(result)-1)
	return result
}

func quickSort(arr []int64, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []int64, low, high int) int {
	pivot := arr[high]
	i := low - 1
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
