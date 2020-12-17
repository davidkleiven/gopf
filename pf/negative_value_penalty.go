package pf

import "math"

// NegativeValuePenalty is a an explicit term that adds a penalty for negative values.
// The function represents
// f(x) = Prefactor*(|x|^Exponent - x^Exponent)
type NegativeValuePenalty struct {
	Prefactor float64
	Exponent int
	Field string
}

// Penalty returns the derivative of the underlying function
func (nvp *NegativeValuePenalty) Penalty(x float64) float64 {
	if x > 0.0 {
		return 0.0
	}
	p := float64(nvp.Exponent)
	return -2.0*nvp.Prefactor*p*math.Pow(x, p-1.0)
}

// Evaluate can be registered as a function in any model
func (nvp *NegativeValuePenalty) Evaluate(i int, bricks map[string]Brick) complex128 {
	return complex(nvp.Penalty(real(bricks[nvp.Field].Get(i))), 0.0)
}

// NewDefaultNegativeValuePenalty returns a new constraint type using the parameters from
//
// Chan, P.Y., Goldenfeld, N. and Dantzig, J., 2009. 
// Molecular dynamics on diffusive time scales from the phase-field-crystal equation.
// Physical Review E, 79(3), p.035701.
func NewDefaultNegativeValuePenalty(field string) NegativeValuePenalty {
	return NegativeValuePenalty{
		Prefactor: -1500.0,
		Exponent: 3,
		Field: field,
	}
}