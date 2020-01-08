package elasticity

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
	"gonum.org/v1/gonum/mat"
)

func TestDilatationalMisfits(t *testing.T) {
	N := 64
	poisson := 0.3
	bulkMod := 50.0
	shear := Shear(bulkMod, poisson)
	matProp := Isotropic(bulkMod, poisson)
	eps := 0.05
	misfit := mat.NewDense(3, 3, []float64{eps, 0.0, 0.0, 0.0, eps, 0.0, 0.0, 0.0, eps})

	for tIdx, test := range []struct {
		a float64
		b float64
		c float64
	}{
		{
			a: 10.0,
			b: 10.0,
			c: 10.0,
		},
		{
			a: 15.0,
			b: 7.0,
			c: 15.0,
		},
		{
			a: 7.0,
			b: 7.0,
			c: 15.0,
		},
	} {
		voxels := Ellipsoid(N, test.a, test.b, test.c)
		origDataVoxels := make([]complex128, len(voxels.Data))
		copy(origDataVoxels, voxels.Data)

		effForce := NewEffectiveForceFromMisfit(matProp, misfit)
		ft := sfft.NewFFT3(N, N, N)
		ft.FFT(voxels.Data)
		force := make([][]complex128, N*N*N)
		for i := range force {
			force[i] = make([]complex128, 3)
		}
		for comp := 0; comp < 3; comp++ {
			forceComp := effForce.Get(comp, ft.Freq, voxels.Data)
			for i := range force {
				force[i][comp] = forceComp[i]
			}
		}

		disp := Displacements(force, ft.Freq, matProp)
		strains := make([]*mat.Dense, len(disp))
		for k := range strains {
			strains[k] = mat.NewDense(3, 3, nil)
		}

		// Calculate strains
		for i := 0; i < 3; i++ {
			for j := i; j < 3; j++ {
				strain := Strain(disp, ft.Freq, i, j)
				ft.IFFT(strain)
				realPart := make([]float64, len(strain))
				for k := range strain {
					realPart[k] = real(strain[k]) / float64(len(strain))
					s := realPart[k]
					if math.Abs(real(origDataVoxels[k])) > 0.5 {
						s -= misfit.At(i, j)
					}
					strains[k].Set(i, j, s)
					strains[k].Set(j, i, s)
				}
			}
		}

		energy := 0.0
		for i := range strains {
			energy += EnergyDensity(matProp, strains[i])
		}
		expect := EshelbyEnergyDensityDilatational(poisson, shear, eps)
		vol := 4.0 * math.Pi * test.a * test.b * test.c / 3.0
		if math.Abs(energy/vol-expect) > 0.05*expect {
			t.Errorf("Relative energy difference in test %d exceeds 0.05. Expected %f got %f", tIdx, expect, energy/vol)
		}
	}
}

func EshelbyEnergyDensityDilatational(poisson float64, shear float64, misfit float64) float64 {
	return 2.0 * (1.0 + poisson) * shear * misfit * misfit / (1.0 - poisson)
}
