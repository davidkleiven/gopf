package pf

import "math"

// TensorialHessian is a type used to represent the term
// K_ij d^2c/dx_idx_j (sum of i and j)
type TensorialHessian struct {
	Field string
	K     []float64
}

// Construct builds the correct RHS term
func (th *TensorialHessian) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		Clear(field)
		brick := bricks[th.Field]
		for i := range field {
			f := freq(i)
			value := brick.Get(i)

			dim := len(f)

			// Diagonal terms
			for j := 0; j < dim; j++ {
				preFactor := -4.0 * math.Pi * math.Pi * f[j] * f[j] * th.GetCoeff(j, j)
				field[i] += complex(preFactor, 0.0) * value
			}

			// Off-diagonal terms
			for j := 0; j < dim; j++ {
				for k := j + 1; k < dim; k++ {
					preFactor := -8.0 * math.Pi * math.Pi * f[j] * f[k] * th.GetCoeff(j, k)
					field[i] += complex(preFactor, 0.0) * value
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
