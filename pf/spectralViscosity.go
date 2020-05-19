package pf

import (
	"math"

	"github.com/davidkleiven/gopf/pfutil"
)

// SpectralViscosity implements the spectral viscosity method proposed in
// Tadmor, E., 1989. Convergence of spectral methods for nonlinear conservation laws.
// SIAM Journal on Numerical Analysis, 26(1), pp.30-44.
// In the fourier transformed domain, the following the spectral viscosity term can be written as
// -eps*Q(k)*k^Power <field>, where field is an arbitrary field name.
// The function Q(k) is an interpolating function defined by. For simplicity m = DissipationThreshold
// is introduced and x = 3k/2m - 1/2
//          **
//          *  0,           if k < m/3
// Q(k) = **   2x^2 - 3x^3, if k >= m/3 and k <= m
//          *  1,           if k > m
//          **
// In the paper by E. Tadmor it is found that Eps*m = 0.25 yields good results
type SpectralViscosity struct {
	Eps                  float64
	DissipationThreshold float64
	Power                int
}

// interpolant smoothly interpolates between 0 at
// 0.33*peakPosition and 1 at peakPosition. 0.33 is chosen because it was
// used in the paper by E. Tadmor
func interpolant(f float64, peakPosition float64) float64 {
	frac := 1.0 / 3.0
	if f < frac*peakPosition {
		return 0.0
	} else if f > peakPosition {
		return 1.0
	}
	x := 1.5 * (f - frac*peakPosition) / peakPosition
	return 2.0*x*x - 3.0*x*x*x
}

// Construct return the denomonator that is required for an implicit treatment
// of the spectral viscosity term
func (sv *SpectralViscosity) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			f := freq(i)
			fRad := math.Sqrt(pfutil.Dot(f, f))
			value := interpolant(fRad, sv.DissipationThreshold)
			field[i] = complex(-sv.Eps*value*math.Pow(fRad, float64(sv.Power)), 0.0)
		}
	}
}

// OnStepFinished empty function implemented to satisfy interface
func (sv *SpectralViscosity) OnStepFinished(t float64, bricks map[string]Brick) {}
