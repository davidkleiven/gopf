package pfutil

import "math"

// CmplxEqualApprox returns true if to complex arrays are equal within the passed tolerance
func CmplxEqualApprox(a []complex128, b []complex128, tol float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if math.Abs(real(a[i])-real(b[i])) > tol || math.Abs(imag(a[i])-imag(b[i])) > tol {
			return false
		}
	}
	return true
}

// ElemwiseAdd adds dst and data and places the result in dst
func ElemwiseAdd(dst []complex128, data []complex128) {
	for i := range dst {
		dst[i] += data[i]
	}
}

// ElemwiseMul multiplies dst and data and places the result in dst.
func ElemwiseMul(dst []complex128, data []complex128) {
	for i := range dst {
		dst[i] *= data[i]
	}
}

// DivRealScalar divides each element in the comlex array by a real scalar
func DivRealScalar(data []complex128, factor float64) []complex128 {
	cfactor := complex(factor, 0.0)
	for i := range data {
		data[i] /= cfactor
	}
	return data
}

// ProdInt calculates the product of all the elements in the passed sequence
func ProdInt(a []int) int {
	res := 1
	for i := range a {
		res *= a[i]
	}
	return res
}

// Dot calculates the dot product between two slices
func Dot(a []float64, b []float64) float64 {
	res := 0.0
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}

// MaxReal calculates the maximum real part
func MaxReal(data []complex128) float64 {
	maxval := real(data[0])
	for i := range data {
		if real(data[i]) > maxval {
			maxval = real(data[i])
		}
	}
	return maxval
}

// MinReal returns the minimum real-part value
func MinReal(data []complex128) float64 {
	minval := real(data[0])
	for i := range data {
		if real(data[i]) < minval {
			minval = real(data[i])
		}
	}
	return minval
}

// Clear sets all elements in the slice to zero
func Clear(data []complex128) {
	for i := range data {
		data[i] = complex(0.0, 0.0)
	}
}

// Wrap wrap pos such that it is inside the box defined by domainsize
func Wrap(pos []int, domainSize []int) {
	for i := range pos {
		if pos[i] < 0 {
			factor := -pos[i] / domainSize[i]
			pos[i] += (factor + 1) * domainSize[i]
		}
		pos[i] = pos[i] % domainSize[i]
	}
}
