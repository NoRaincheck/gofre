package core

import "C"

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

//export MatrixTranspose
func MatrixTranspose(matrix []float64, rows, cols int64) []float64 {
	result := make([]float64, rows*cols)
	
	for i := int64(0); i < rows; i++ {
		for j := int64(0); j < cols; j++ {
			result[j*rows+i] = matrix[i*cols+j]
		}
	}
	
	return result
}

//export DotProduct
func DotProduct(a, b []float64) float64 {
	sum := 0.0
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}
