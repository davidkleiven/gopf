package pf

import (
	"math"

	"github.com/davidkleiven/gopf/pfutil"
)

// ChargeTransport implements a term corresponding to the rate of change
// of local charge density in an electric field. The implementation follows
// closely the approach in
//
// Jin, Y.M., 2013. Phase field modeling of current density distribution and
// effective electrical conductivity in complex microstructures.
// Applied Physics Letters, 103(2), p.021906.
// https://doi.org/10.1063/1.4813392
//
// The evolution of the charge density is described by
//
// dp/dt = -div j, where p is the charge density and j is the current density
// given by j_i = sigma_ij E_j, where sigma is the conductivity tensor and E
// is the local electric field. The local electriv field is calculated by solving
// poisson equation
// LAP phi = -p/eps_0, eps_0 is the vacuum permittivity
// The net result of this is that the local electric field is given by
//                      **
//                i    *   d^3 k      p(k)k
// E = E_ext - ------- * ---------  -------- e^(ik*r)
//              eps_0  *  (2pi)^3      k^2
//                   **
// where p(k) is the fourier transform of the density, k is the reciprocal wave vector
// and r is the real space position vector. E_ext is the external electric field
type ChargeTransport struct {
	// Conductivity return the conductivity tensor at node i. If 3D
	// the order should be s_xx, s_yy, s_zz, s_xz, s_yz, s_xy and if 2D
	// it should be s_xx, s_yy, s_xy. The vacuum permittivity should be embeded
	// in the conductivity. Thus, if the "normal" conductiviy in Ohm's law is
	// labeled sigma, the current density is given by j = sigma*E, this function
	// should return sigma/eps_0, where eps_0 is the vacumm permittivity
	Conductivity func(i int) []float64

	// ExternalField represents the external electric field
	ExternalField []float64

	// Field is the name of the field that corresponds to the charge density
	Field string

	// FT is a fourier transformer
	FT FourierTransform
}

// current calculate the current. The Current will be given in the passed
// effective current array. The length of effCurrent must be at least dim*N
// where N is the number of nodes. brick is the brick the represents the charge
// density
func (ct *ChargeTransport) current(brick Brick, N int) []complex128 {
	dim := len(ct.FT.Freq(0))
	workArray := make([]complex128, (1+dim)*N)
	effField := workArray[:N]
	effCurrent := workArray[N:]

	for d := 0; d < dim; d++ {
		// Calculate the d-th component of the effective electric field
		// (external + induced)
		for i := 0; i < N; i++ {
			kVec := ct.FT.Freq(i)
			kSq := pfutil.Dot(kVec, kVec)
			if math.Abs(math.Abs(kVec[d])-0.5) > 1e-10 {
				effField[i] = brick.Get(i) * complex(0.0, kVec[d]/(2.0*math.Pi*kSq+1e-16))
			} else {
				effField[i] = 0.0
			}

		}
		effField = ct.FT.IFFT(effField)
		pfutil.DivRealScalar(effField, float64(N))
		for i := range effField {
			effField[i] -= complex(ct.ExternalField[d], 0.0)
		}

		// Update the effective current by adding the contribution from the d-th component
		// of the electric field
		for i := 0; i < N; i++ {
			sigma := ct.Conductivity(i)
			for d2 := 0; d2 < dim; d2++ {
				start := d2 * N
				effCurrent[start+i] += complex(sigma[voigtIndex(d, d2, dim)], 0.0) * effField[i]
			}
		}
	}
	return effCurrent
}

// Construct returns the fourier transformed value of
func (ct *ChargeTransport) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		pfutil.Clear(field)
		dim := len(freq(0))

		brick := bricks[ct.Field]
		effCurrent := ct.current(brick, len(field))
		work := make([]complex128, len(field))

		// Fourier transform the currents
		for d2 := 0; d2 < dim; d2++ {
			start := d2 * len(field)
			end := (d2 + 1) * len(field)
			work = ct.FT.FFT(effCurrent[start:end])

			// Update the divergence of the current
			for i := range field {
				k := freq(i)[d2]
				if math.Abs(math.Abs(k)-0.5) > 1e-10 {
					field[i] += complex(0.0, 2.0*math.Pi*k) * work[i]
				}
			}
		}
	}
}

// Current returns the current in real space. Density should give the fourier transformed
// or real space density. If it represents the realspace density, the realspace flag should
// be set to true (in which case a fourier transform is performed internally) otherwise it
// should be set to false
// charge density
func (ct *ChargeTransport) Current(density Brick, N int, realspace bool) [][]float64 {
	if realspace {
		rdensity := make([]complex128, N)
		for i := 0; i < N; i++ {
			rdensity[i] = density.Get(i)
		}
		rdensity = ct.FT.FFT(rdensity)
		density = NewField("ftDensity", N, rdensity)
	}

	current := ct.current(density, N)
	dim := len(ct.FT.Freq(0))
	res := make([][]float64, dim)
	for d := 0; d < dim; d++ {
		res[d] = make([]float64, N)
		for i := 0; i < N; i++ {
			res[d][i] = -real(current[d*N+i])
		}
	}
	return res
}

// OnStepFinished does nothing for this term
func (ct *ChargeTransport) OnStepFinished(t float64, b map[string]Brick) {}

func voigtIndex3D() [][]int {
	return [][]int{
		{0, 5, 4},
		{5, 1, 3},
		{4, 3, 2},
	}
}

func voigtIndex2D() [][]int {
	return [][]int{
		{0, 2},
		{2, 1},
	}
}

func voigtIndex(i, j, dim int) int {
	if dim == 2 {
		return voigtIndex2D()[i][j]
	}
	return voigtIndex3D()[i][j]
}
