package elasticity

import (
	"math"
	"testing"
)

func TestStrainEnergyCLI(t *testing.T) {
	N := 64
	poisson := 0.3
	bulkMod := 50.0
	shear := Shear(bulkMod, poisson)
	eps := 0.05

	c11 := bulkMod + 4.0*shear/3.0
	c12 := bulkMod - 2.0*shear/3.0
	params := StrainEnergyInput{
		HalfA: 8.0,
		HalfB: 8.0,
		HalfC: 8.0,
		MatPropMatrix: []float64{c11, c12, c12, 0.0, 0.0, 0.0,
			c12, c11, c12, 0.0, 0.0, 0.0,
			c12, c12, c11, 0.0, 0.0, 0.0,
			0.0, 0.0, 0.0, shear, 0.0, 0.0,
			0.0, 0.0, 0.0, 0.0, shear, 0.0,
			0.0, 0.0, 0.0, 0.0, 0.0, shear},
		Misfit:     []float64{eps, eps, eps, 0.0, 0.0, 0.0},
		DomainSize: N,
	}

	energy := CalculateStrainEnergy(params)
	expect := EshelbyEnergyDensityDilatational(poisson, shear, eps)

	tol := 1e-4
	if math.Abs(energy-expect) > tol {
		t.Errorf("Strain energy CLI: expected %f got %f", expect, energy)
	}
}
