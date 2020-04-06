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

// GetEnergy evaluates the energy contribution from this term. The fields
// in bricks should be the real space representation. The number of nodes
// in the simulation cell is specified in the nodes argument
func (pct *PairCorrlationTerm) GetEnergy(bricks map[string]Brick, ft FourierTransform, domainSize []int) float64 {
	numNodes := ProdInt(domainSize)
	b := bricks[pct.Field]
	field := make([]complex128, numNodes)
	for i := 0; i < numNodes; i++ {
		field[i] = b.Get(i)
	}
	ft.FFT(field)
	for i := range field {
		f := ft.Freq(i)
		fRad := math.Sqrt(Dot(f, f))
		field[i] *= complex(pct.Prefactor*pct.PairCorrFunc.Eval(2.0*math.Pi*fRad), 0.0)
	}
	ft.IFFT(field)
	DivRealScalar(field, float64(numNodes))

	integral := 0.0
	for i := range field {
		value := real(field[i] * b.Get(i))
		integral += value
	}
	return -0.5 * integral
}

// IdealMixtureTerm implements the ideal mixture model used in the paper by Greenwood et al.
// To use this term in a model, register the function Eval as a function in the model.
// Prefactor is a constant factor that is multiplied with the energy
type IdealMixtureTerm struct {
	IdealMix  pfc.IdealMix
	Field     string
	Prefactor float64
}

// Eval returns the negative derivative of the underlying ideal mixture term
func (idt *IdealMixtureTerm) Eval(i int, bricks map[string]Brick) complex128 {
	value := real(bricks[idt.Field].Get(i))
	return complex(-idt.Prefactor*idt.IdealMix.Deriv(value), 0.0)
}

// GetEnergy evaluates the energy contribution from this term. The fields
// in bricks should be the real space representation. The number of nodes
// in the simulation cell is specified in the nodes argument
func (idt *IdealMixtureTerm) GetEnergy(bricks map[string]Brick, nodes int) float64 {
	b := bricks[idt.Field]
	integral := 0.0
	for i := 0; i < nodes; i++ {
		val := real(b.Get(i))
		integral += idt.Prefactor * idt.IdealMix.Eval(val)
	}
	return integral
}
