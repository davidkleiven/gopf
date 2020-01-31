package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

func TestGradientCalculator(t *testing.T) {
	N := 16
	data := make([]complex128, N*N)
	expect := make([]float64, N*N)
	for i := range data {
		x := float64(i%N) / float64(N)
		data[i] = complex(x*x-2*x*x*x+x*x*x*x, 0.0)
		v := 2.0*x - 6.0*x*x + 4.0*x*x*x
		expect[i] = v / float64(N)
	}

	ft := sfft.NewFFT2(N, N)
	grad := GradientCalculator{
		FT:   ft,
		Comp: 1,
	}

	got := make([]complex128, N*N)
	grad.Calculate(data, got)
	tol := 1e-4
	for i := range got {
		re := real(got[i])
		im := imag(got[i])
		if math.Abs(re-expect[i]) > tol || math.Abs(im) > tol {
			diff := re - expect[i]
			t.Errorf("Expected (%f, 0) got (%f, %f). Real part diff %f\n", expect[i], re, im, diff)
		}
	}
}
