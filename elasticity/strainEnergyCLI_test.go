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

func TestEnergyWithPerturbation(t *testing.T) {
	N := 32
	poisson := 0.3
	bulkMod := 50.0
	shear := Shear(bulkMod, poisson)
	eps := 0.05

	c11 := bulkMod + 4.0*shear/3.0
	c12 := bulkMod - 2.0*shear/3.0
	fvals := []float64{0.2, 0.4, 0.6, 0.8, 1.0}
	energies := make([]float64, len(fvals))
	for i := range fvals {
		f := fvals[i]
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
			Misfit:            []float64{eps, eps, eps, 0.0, 0.0, 0.0},
			DomainSize:        N,
			ApplyPerturbation: true,
			MatPropInc: []float64{f * c11, f * c12, f * c12, 0.0, 0.0, 0.0,
				f * c12, f * c11, f * c12, 0.0, 0.0, 0.0,
				f * c12, f * c12, f * c11, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, f * shear, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, f * shear, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, f * shear},
		}
		energy := CalculateStrainEnergy(params)
		energies[i] = energy
	}

	// We expect a linear dependence on f
	diff := energies[1] - energies[0]
	for i := 1; i < len(energies); i++ {
		d := energies[i] - energies[i-1]
		if math.Abs(d-diff) > 0.1*diff {
			t.Errorf("%d: Not a linear relation: %v. Dev. %f tolerates %f \n", i, energies, d-diff, 0.1*diff)
		}
	}

}
