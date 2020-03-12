package pf

import "math"

// Peak represent one peak in the reciprocal space pair correlation function
type Peak struct {
	PlaneDensity float64
	Location     float64
	Width        float64
	NumPlanes    int
}

// ReciprocalSpacePairCorrelation implements the type of correlation function
// presented by Greenwood et al.
//
// Greenwood, M., Provatas, N. and Rottler, J., 2010.
// Free energy functionals for efficient phase field crystal modeling of structural phase transformations.
// Physical review letters, 105(4), p.045702.
type ReciprocalSpacePairCorrelation struct {
	EffTemp float64
	Peaks   []Peak
}

// Eval evaluates the pair correlation function
func (rspc ReciprocalSpacePairCorrelation) Eval(k float64) float64 {
	result := 0.0
	for _, peak := range rspc.Peaks {
		prefactor := math.Exp(-rspc.EffTemp * rspc.EffTemp * k * k / (2.0 * peak.PlaneDensity * float64(peak.NumPlanes)))
		value := prefactor * math.Exp(-0.5*math.Pow((k-peak.Location)/peak.Width, 2))
		if value > result {
			result = value
		}
	}
	return result
}

// SquareLattice2D construct the two peaks with the lowest frequency peaks for the square lattice in 2D.
// The width of the two peaks is given as an array and the lattice parameter is given via a. The unit of
// both width and a is pixels.
func SquareLattice2D(width float64, a float64) []Peak {
	a2 := a / math.Sqrt(2.0)
	return []Peak{
		Peak{
			PlaneDensity: 1.0,
			NumPlanes:    4,
			Width:        width,
			Location:     2.0 * math.Pi / a,
		},
		Peak{
			PlaneDensity: 1.0 / math.Sqrt(2.0),
			NumPlanes:    4,
			Width:        width,
			Location:     2.0 * math.Pi / a2,
		},
	}
}

// TriangularLattice2D returns the peaks nessecary to describe a triangular lattice in 2D
func TriangularLattice2D(width float64, a float64) []Peak {
	return []Peak{
		Peak{
			PlaneDensity: 2.0,
			NumPlanes:    3,
			Width:        width,
			Location:     2.0 * math.Pi / a,
		},
	}
}

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
	PairCorrFunc ReciprocalSpacePairCorrelation
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
