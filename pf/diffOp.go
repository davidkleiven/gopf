package pf

import (
	"math"

	"gonum.org/v1/gonum/floats"
)

// Frequency is a function that returned the frequency at position i
type Frequency func(i int) []float64

// DiffOp is a differential operator. It takes an object that has the
// frequency method implemented and applies the operator in-place to
// the passed array
type DiffOp func(freq Frequency, ft []complex128) []complex128

// LaplacianN is a type used for the Laplacian operator raised to some power
type LaplacianN struct {
	Power int
}

// Eval implements the fourier transformed laplacian operator. freq is a function
// that returns the frequency at index i of the passed array. ft is the fourier transformed
// field
func (l LaplacianN) Eval(freq Frequency, ft []complex128) []complex128 {
	for i := range ft {
		ft[i] *= complex(math.Pow(-math.Pow(2.0*math.Pi*floats.Norm(freq(i), 2), 2.0), float64(l.Power)), 0.0)
	}
	return ft
}
