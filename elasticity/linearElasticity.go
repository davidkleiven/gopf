package elasticity

import (
	"math"

	"github.com/davidkleiven/gopf/pfutil"
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

// HomogeneousModulusEnergy returns the elastic energy from a
func HomogeneousModulusEnergy(indicator []complex128, domainSize []int, misfit *mat.Dense, matProp Rank4) float64 {
	volume := 0.0
	for i := range indicator {
		volume += real(indicator[i])
	}

	effForce := NewEffectiveForceFromMisfit(matProp, misfit)
	ft := pfutil.NewFFTW(domainSize)
	ft.FFT(indicator)
	force := make([][]complex128, len(indicator))
	for k := range force {
		force[k] = make([]complex128, 3)
	}

	for comp := 0; comp < 3; comp++ {
		fComp := effForce.Get(comp, ft.Freq, indicator)
		for k := range force {
			force[k][comp] = fComp[k]
		}
	}

	// Wrap the freq function in case of 2D calculation
	freq := func(i int) []float64 {
		if len(domainSize) == 3 {
			return ft.Freq(i)
		}
		f3 := make([]float64, 3)
		f := ft.Freq(i)
		copy(f3[:2], f)
		return f3
	}

	disp := Displacements(force, freq, &matProp)
	// Inservse FFT such that we can use it distinguish regions
	ft.IFFT(indicator)
	for i := range indicator {
		indicator[i] /= complex(float64(len(indicator)), 0.0)
	}

	energy := 0.0
	strains := make([]*mat.Dense, len(disp))
	for k := range strains {
		strains[k] = mat.NewDense(3, 3, nil)
	}

	for i := 0; i < 3; i++ {
		for j := i; j < 3; j++ {
			strain := Strain(disp, freq, i, j)
			ft.IFFT(strain)
			for k := range strain {
				re := real(strain[k]) / float64(len(strain))

				if real(indicator[k]) > 0.5 {
					re -= misfit.At(i, j)
				}
				strains[k].Set(i, j, re)
				strains[k].Set(j, i, re)
			}
		}
	}
	for i := range strains {
		energy += EnergyDensity(&matProp, strains[i])
	}
	return energy / volume
}
