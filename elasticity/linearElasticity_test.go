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

		disp := Displacements(force, ft.Freq, &matProp)
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
			energy += EnergyDensity(&matProp, strains[i])
		}
		expect := EshelbyEnergyDensityDilatational(poisson, shear, eps)
		vol := 4.0 * math.Pi * test.a * test.b * test.c / 3.0
		if math.Abs(energy/vol-expect) > 0.05*expect {
			t.Errorf("Relative energy difference in test %d exceeds 0.05. Expected %f got %f", tIdx, expect, energy/vol)
		}
	}
}

func TestStrain(t *testing.T) {
	N := 8
	ux := make([]complex128, N*N)
	expectDeriv := make([]float64, N*N)
	for i := range ux {
		row := i / N
		x := float64(row) / float64(N)

		ux[i] = complex(math.Pow(x*(1.0-x), 2), 0.0)
		expectDeriv[i] = 2.0*x*(1-x)*(1-x) - 2*x*x*(1-x)
		expectDeriv[i] /= float64(N)
	}

	ft := sfft.NewFFT2(N, N)
	ft.FFT(ux)

	ftDisplacement := make([][]complex128, N*N)
	for i := range ftDisplacement {
		ftDisplacement[i] = make([]complex128, 2)
		ftDisplacement[i][0] = ux[i]
	}
	zeros := make([]float64, N*N)
	for tnum, test := range []struct {
		i      int
		j      int
		expect []float64
	}{
		{
			i:      0,
			j:      0,
			expect: expectDeriv,
		},
		{
			i:      0,
			j:      1,
			expect: zeros,
		},
		{
			i:      1,
			j:      1,
			expect: zeros,
		},
	} {
		strain := Strain(ftDisplacement, ft.Freq, test.i, test.j)
		ft.IFFT(strain)
		tol := 1e-3
		for i := range strain {
			re := real(strain[i]) / float64(len(strain))
			im := imag(strain[i]) / float64(len(strain))

			if math.Abs(re-test.expect[i]) > tol || math.Abs(im) > tol {
				t.Errorf("Test #%d: Expected (%f,0) got (%f, %f)\n", tnum, test.expect[i], re, im)
			}
		}
	}

}

func EshelbyEnergyDensityDilatational(poisson float64, shear float64, misfit float64) float64 {
	return 2.0 * (1.0 + poisson) * shear * misfit * misfit / (1.0 - poisson)
}
