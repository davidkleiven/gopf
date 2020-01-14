package elasticity

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// GridFunc is a function type that returns value from a grid
type GridFunc func(i int) float64

// FourierTransform is a generic interface for a multidimensional FFT
type FourierTransform interface {
	Freq(i int) []float64

	FFT(data []complex128) []complex128

	IFFT(data []complex128) []complex128
}

// PerturbedForce returns the effective force for the first order perturbation arising from
// difference in the material properties. The deviation in material properties at position i
// is given by shape(i)*matProp. The fourier transformed effective force arising from the perturbation is given by
//
// -delta C_{ijkl}*i*k_j(FT(misfit_kl - zero_eps_kl)shape(i))
//
// where FT denotes the fourier transform, misfit_kl denotes the kl component of the misfit strain,
// zero_eps_kl denotes the kl component of the zeroth order strain (e.g. strain from the case where the
// elastic properties are constant)
func PerturbedForce(ft FourierTransform, misfit *mat.Dense, ftDisp [][]complex128, shape GridFunc, matProp Rank4Tensor) [][]complex128 {
	force := make([][]complex128, len(ftDisp))
	for i := range force {
		force[i] = make([]complex128, 3)
	}

	dim := len(ft.Freq(0))

	for i := 0; i < dim; i++ {
		for j := i; j < dim; j++ {
			factor := 1.0
			if i != j {
				factor = 2.0
			}
			strain := Strain(ftDisp, ft.Freq, i, j)
			ft.IFFT(strain)

			// Multiply with the shape function
			for k := range strain {
				v := (misfit.At(i, j) - real(strain[k])/float64(len(strain))) * shape(k)
				strain[k] = complex(v, 0.0)
			}
			ft.FFT(strain)

			for k := range force {
				f := ft.Freq(k)
				for comp := 0; comp < dim; comp++ {
					for q := 0; q < 3; q++ {
						force[k][comp] -= complex(0.0, 2.0*math.Pi*factor*matProp.At(comp, q, i, j)*f[comp]) * strain[k]
					}
				}
			}
		}
	}
	return force
}
