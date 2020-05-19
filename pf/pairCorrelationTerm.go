package pf

import (
	"fmt"
	"math"

	"github.com/davidkleiven/gopf/pfc"
	"github.com/davidkleiven/gopf/pfutil"
)

// PairCorrlationTerm implements the functional deriviative with respect to Q of the functional
//                  **        **
//             A   *         *
//  g[Q] = -  ---  * dr Q(r) * dr' C(|r - r'|)Q(r')
//             2   *         *
//               **        **
// where C(|r-r'|) is a pair correlation function. PairCorrFunc is the fourier transform of the
// pair correlation function and Field is the name of the field (e.g. name of Q in the equation above).
// Prefactor is a constant factor that is multiplied with the integral (A in the equation above).
// The attribute Laplacian determines whether the the laplacian should be applied to the functional
// derivative or not. If true, then this term represents nabla^2 dg/dQ, otherwise it simply represents
// dg/dQ. When used in a model this term should be registered as an implicit term
type PairCorrlationTerm struct {
	PairCorrFunc pfc.ReciprocalSpacePairCorrelation
	Field        string
	Prefactor    float64
	Laplacian    bool
}

// evaluate returns the prefactor times the correlation function evaluated
// at the passed frequency
func (pct *PairCorrlationTerm) evaluate(f float64) float64 {
	return pct.Prefactor * pct.PairCorrFunc.Eval(2.0*math.Pi*f)
}

// Construct builds the rhs required to represent the term
func (pct *PairCorrlationTerm) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, out []complex128) {
		for i := range out {
			f := freq(i)
			fRad := math.Sqrt(pfutil.Dot(f, f))
			out[i] = -complex(pct.evaluate(fRad), 0.0)
		}

		if pct.Laplacian {
			lap := LaplacianN{Power: 1}
			lap.Eval(freq, out)
		}
	}
}

// OnStepFinished is simply included to satisfy the UserDefinedTerm interface
func (pct *PairCorrlationTerm) OnStepFinished(t float64, bricks map[string]Brick) {}

// GetEnergy evaluates the energy contribution from this term. The fields
// in bricks should be the real space representation. The number of nodes
// in the simulation cell is specified in the nodes argument
func (pct *PairCorrlationTerm) GetEnergy(bricks map[string]Brick, ft FourierTransform, domainSize []int) float64 {
	numNodes := pfutil.ProdInt(domainSize)
	b := bricks[pct.Field]
	field := make([]complex128, numNodes)
	for i := 0; i < numNodes; i++ {
		field[i] = b.Get(i)
	}
	ft.FFT(field)
	for i := range field {
		f := ft.Freq(i)
		fRad := math.Sqrt(pfutil.Dot(f, f))
		field[i] *= complex(pct.Prefactor*pct.PairCorrFunc.Eval(2.0*math.Pi*fRad), 0.0)
	}
	ft.IFFT(field)
	pfutil.DivRealScalar(field, float64(numNodes))

	integral := 0.0
	for i := range field {
		value := real(field[i] * b.Get(i))
		integral += value
	}
	return -0.5 * integral
}

// ExplicitPairCorrelationTerm implement the pair correlation function, but the construct method
// returns the expression corresponding to an explicit treatment of the term in the PDE.
// The only difference between ExplicitPairCorrelationTerm and PairCorrelationTerm is that the
// Construct method returns the explicit and implicit variant, respectively.
type ExplicitPairCorrelationTerm struct {
	PairCorrlationTerm
}

// Construct returns a function that evaluates the RHS of the PDE
func (epct *ExplicitPairCorrelationTerm) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, out []complex128) {
		brick := bricks[epct.Field]
		for i := range out {
			f := freq(i)
			fRad := math.Sqrt(pfutil.Dot(f, f))
			out[i] = -complex(epct.evaluate(fRad), 0.0) * brick.Get(i)
		}

		if epct.Laplacian {
			lap := LaplacianN{Power: 1}
			lap.Eval(freq, out)
		}
	}
}

// IdealMixtureTerm implements the ideal mixture model used in the paper by Greenwood et al.
// To use this term in a model, register the function Eval as a function in the model.
// Prefactor is a constant factor that is multiplied with the energy. When used together with
// a model, this term should be registered as a mixed term. Laplacian indicates whether the
// laplacian should be applied to the term. If true then the laplacian will be applied internally
type IdealMixtureTerm struct {
	IdealMix  pfc.IdealMix
	Field     string
	Prefactor float64
	Laplacian bool
}

// Eval returns the derivative of the underlying ideal mixture term
func (idt *IdealMixtureTerm) Eval(i int, bricks map[string]Brick) complex128 {
	value := real(bricks[idt.Field].Get(i))
	return complex(idt.Prefactor*idt.IdealMix.Deriv(value), 0.0)
}

// ConstructLinear builds the linear part
func (idt *IdealMixtureTerm) ConstructLinear(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			field[i] = complex(idt.Prefactor, 0.0)
		}

		if idt.Laplacian {
			lap := LaplacianN{Power: 1}
			lap.Eval(freq, field)
		}
	}
}

func (idt *IdealMixtureTerm) nonLinearDerivedFieldName() string {
	return fmt.Sprintf("ideal_mixture_%s_nonlin", idt.Field)
}

// DerivedField returns the required derived field that is nessecary in order
// to use this model
func (idt *IdealMixtureTerm) DerivedField(numNodes int, bricks map[string]Brick) DerivedField {
	return DerivedField{
		Name: idt.nonLinearDerivedFieldName(),
		Data: make([]complex128, numNodes),
		Calc: func(out []complex128) {
			b := bricks[idt.Field]
			for i := range out {
				v := real(b.Get(i))
				out[i] = complex(3.0*idt.IdealMix.ThirdOrderPrefactor()*v*v+4.0*idt.IdealMix.FourthOrderPrefactor()*v*v*v, 0.0)
			}
		},
	}
}

// ConstructNonLinear builds the non linear part
func (idt *IdealMixtureTerm) ConstructNonLinear(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		if b, ok := bricks[idt.nonLinearDerivedFieldName()]; ok {
			for i := range field {
				field[i] = b.Get(i)
			}

			if idt.Laplacian {
				lap := LaplacianN{Power: 1}
				lap.Eval(freq, field)
			}
		} else {
			msg := fmt.Sprintf("Missing derived field %s.\n", idt.nonLinearDerivedFieldName())
			msg += "Make sure that the field returned by IdealMixtureTerm.DerivedField is registered"
			panic(msg)
		}
	}
}

// OnStepFinished does nothing
func (idt *IdealMixtureTerm) OnStepFinished(t float64, bricks map[string]Brick) {}

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
