package pf

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
