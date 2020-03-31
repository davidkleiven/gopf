package pf

import (
	"math"

	"github.com/davidkleiven/gopf/pfc"
)

// PairCorrlationTerm implements the functional deriviative with respect to Q of the functional
//               **        **
//          A   *         *
//  g[Q] = ---  * dr Q(r) * dr' C(|r - r'|)Q(r')
//          2   *         *
//            **        **
// where C(|r-r'|) is a pair correlation function. PairCorrFunc is the fourier transform of the
// pair correlation function and Field is the name of the field (e.g. name of Q in the equation above).
// Prefactor is a constant factor that is multiplied with the integral (A in the equation above)
type PairCorrlationTerm struct {
	PairCorrFunc pfc.ReciprocalSpacePairCorrelation
	Field        string
	Prefactor    float64
}

// Construct builds the rhs required to represent the term
func (pct *PairCorrlationTerm) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, out []complex128) {
		brick := bricks[pct.Field]
		for i := range out {
			f := freq(i)
			fRad := math.Sqrt(Dot(f, f))
			out[i] = complex(pct.Prefactor*pct.PairCorrFunc.Eval(2.0*math.Pi*fRad), 0.0) * brick.Get(i)
		}
	}
}
