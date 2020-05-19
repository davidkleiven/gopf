package pf

import (
	"math"

	"github.com/davidkleiven/gopf/pfutil"
)

// BrickPlaceholder is struct that is used as a brick, in case no brick is specified
type BrickPlaceholder struct{}

// Get return 1.0
func (bp *BrickPlaceholder) Get(i int) complex128 {
	return complex(1.0, 0.0)
}

// TensorialHessian is a type used to represent the term
// K_ij d^2c/dx_idx_j (sum of i and j). The tensial hessian brick
// can be added to an equation in two ways as it can be treated both
// explicitly and implicitly (in some cases). If the term can be
// treated implicitly, is recommended that one specify it such that the
// equation parser understands that it should be dealt with implicitly.
// Cases where it can be treated implicitly is if the hessian should be
// applied to the same field as located on the left hand side
//
// d<field>/dt = K_ij d^2<field>/dx_idx_j
//
// Note that this term can only be applied to the field that is also on the
// left hand side of the equation. The following will not work
//
// d<fieldA>/dt = K_ij d^2<fieldB>/dx_idx_j
type TensorialHessian struct {
	Field string
	K     []float64
}

// Construct builds the correct RHS term
func (th *TensorialHessian) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		pfutil.Clear(field)

		for i := range field {
			f := freq(i)

			dim := len(f)

			// Diagonal terms
			for j := 0; j < dim; j++ {
				preFactor := -4.0 * math.Pi * math.Pi * f[j] * f[j] * th.GetCoeff(j, j)
				field[i] += complex(preFactor, 0.0)
			}

			// Off-diagonal terms
			for j := 0; j < dim; j++ {
				for k := j + 1; k < dim; k++ {
					preFactor := -8.0 * math.Pi * math.Pi * f[j] * f[k] * th.GetCoeff(j, k)
					field[i] += complex(preFactor, 0.0)
				}
			}
		}
	}
}

// GetCoeff return the i, j element of the coefficient tensor
func (th *TensorialHessian) GetCoeff(i, j int) float64 {
	d := 2
	if len(th.K) == 9 {
		d = 3
	}
	return th.K[i*d+j]
}

// OnStepFinished does nothing as there is no need for updates for this term
func (th *TensorialHessian) OnStepFinished(t float64, bricks map[string]Brick) {}
