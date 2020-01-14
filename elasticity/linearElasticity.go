package elasticity

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// Frequency is a function that can return the frequency corresponding to
// an index in a 1D array (See the SFFT package)
type Frequency func(i int) []float64

// DisplacementMatrixElement returns the (m, n) element of the matrix needed to
// find the displacements
func DisplacementMatrixElement(m, n int, freq []float64, matProp Rank4Tensor) float64 {
	elem := 0.0
	for j := 0; j < 3; j++ {
		for l := 0; l < 3; l++ {
			elem += matProp.At(m, j, n, l) * freq[j] * freq[l]
		}
	}
	return elem * math.Pow(2.0*math.Pi, 2)
}

// Rank4Tensor is an interface of entities that implementes an getter
// with four indices
type Rank4Tensor interface {
	At(i, j, k, l int) float64
}

// Displacements calculates the fourier transformed displacements from a given body force
func Displacements(ftBodyForce [][]complex128, freq Frequency, matProp Rank4Tensor) [][]complex128 {
	matrix := mat.NewDense(3, 3, nil)
	bodyF := mat.NewDense(3, 2, nil)
	disp := mat.NewDense(3, 2, nil)
	cdisp := make([][]complex128, len(ftBodyForce))
	tol := 1e-10
	for i := range ftBodyForce {
		f := freq(i)

		if math.Abs(f[0]) < tol && math.Abs(f[1]) < tol && math.Abs(f[2]) < tol {
			cdisp[i] = make([]complex128, 3)
			continue
		}
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				matrix.Set(j, k, DisplacementMatrixElement(j, k, f, matProp))
			}
		}

		for j := range ftBodyForce[i] {
			bodyF.Set(j, 0, real(ftBodyForce[i][j]))
			bodyF.Set(j, 1, imag(ftBodyForce[i][j]))
		}
		disp.Solve(matrix, bodyF)
		cdisp[i] = make([]complex128, 3)
		for j := range ftBodyForce[i] {
			cdisp[i][j] = complex(disp.At(j, 0), disp.At(j, 1))
		}
	}
	return cdisp
}

// Strain returns the fourier transformed strains calculated from the
// fourier transformed displacements
func Strain(ftDisp [][]complex128, freq Frequency, m, n int) []complex128 {
	s := make([]complex128, len(ftDisp))
	tol := 1e-10
	for i := range ftDisp {
		f := freq(i)
		fm := f[m]
		fn := f[n]
		if math.Abs(math.Abs(fm)-0.5) < tol {
			fm = 0.0
		}
		if math.Abs(math.Abs(fn)-0.5) < tol {
			fn = 0.0
		}
		s[i] = complex(0.0, math.Pi*fn)*ftDisp[i][m] + complex(0.0, math.Pi*fm)*ftDisp[i][n]
	}
	return s
}

// EnergyDensity calculates the strain energy
func EnergyDensity(matProp Rank4Tensor, strain *mat.Dense) float64 {
	res := 0.0
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					res += matProp.At(i, j, k, l) * strain.At(i, j) * strain.At(k, l)
				}
			}
		}
	}
	return 0.5 * res
}
