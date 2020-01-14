package elasticity

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
	"gonum.org/v1/gonum/mat"
)

func TestPerturbation(t *testing.T) {
	N := 8
	zerothDisp := make([][]complex128, N*N)
	ft := sfft.NewFFT2(N, N)
	ux := make([]complex128, N*N)
	for i := 0; i < N*N; i++ {
		row := i / N
		x := float64(row) / float64(N)
		ux[i] = complex(math.Sin(2.0*math.Pi*x), 0.0)
	}
	ft.FFT(ux)
	for i := range ux {
		zerothDisp[i] = make([]complex128, 3)
		zerothDisp[i][0] = ux[i]
	}

	shpFunc := func(i int) float64 {
		col := i % N
		y := float64(col) / float64(N)
		return math.Sin(2.0 * math.Pi * y)
	}

	matProp := CubicMaterial(100.0, 60.0, 30.0)

	misfit := mat.NewDense(3, 3, []float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0})
	force := PerturbedForce(ft, misfit, zerothDisp, shpFunc, &matProp)

	tmpForce := make([]complex128, len(force))
	tol := 1e-8
	for comp := 0; comp < 2; comp++ {
		for k := range force {
			tmpForce[k] = force[k][comp]
		}
		ft.IFFT(tmpForce)

		for k := range tmpForce {
			row := k / N
			col := k % N
			x := float64(row) / float64(N)
			y := float64(col) / float64(N)
			expect := 4.0 * math.Pi * math.Pi * (matProp.At(comp, 1, 0, 0)*math.Cos(2.0*math.Pi*x)*math.Cos(2.0*math.Pi*y) -
				matProp.At(comp, 0, 0, 0)*math.Sin(2.0*math.Pi*x)*math.Sin(2.0*math.Pi*y)) / float64(N*N)

			res := real(tmpForce[k]) / float64(len(tmpForce))

			if math.Abs(res-expect) > tol {
				t.Errorf("Comp #%d: Expected %f got %f", comp, expect, res)
			}
		}
	}
}
