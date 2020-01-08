package elasticity

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// EffectiveForce is a type that is used to calculate an effective body force
// from a region of misfit strains
type EffectiveForce struct {
	EffStress *mat.Dense
}

// NewEffectiveForceFromMisfit returns the effective force based on the elastic properties
// and the misfit strains
func NewEffectiveForceFromMisfit(matProp Rank4, misfit *mat.Dense) EffectiveForce {
	var eff EffectiveForce
	eff.EffStress = matProp.ContractLast(misfit)
	return eff
}

// Get a component of the effective force specified by comp. freq is a function that
// can returns the frequency of node i. indicator is a fourier transformed indicator function
// of the domain where the misfit strain exists.
func (e *EffectiveForce) Get(comp int, freq Frequency, indicator []complex128) []complex128 {
	force := make([]complex128, len(indicator))
	for i := range indicator {
		k := freq(i)
		for j := range k {
			force[i] += complex(0.0, -e.EffStress.At(comp, j)*2.0*math.Pi*k[j]) * indicator[i]
		}
	}
	return force
}
