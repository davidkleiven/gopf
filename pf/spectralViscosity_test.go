package pf

import (
	"math"
	"testing"
)

func TestInterpolation(t *testing.T) {
	peak := 0.6
	tol := 1e-10
	for i, test := range []struct {
		k      float64
		expect float64
	}{
		{
			k:      0.19,
			expect: 0.0,
		},
		{
			k:      0.66,
			expect: 1.0,
		},
		{
			k:      0.4,
			expect: 1.0 / 8.0,
		},
	} {
		res := interpolant(test.k, peak)
		if math.Abs(res-test.expect) > tol {
			t.Errorf("interpolate test #%d: Expected %f got %f\n", i, test.expect, res)
		}
	}
}

func TestSpectralViscosityTerm(t *testing.T) {
	spectral := SpectralViscosity{
		Eps:                  1.0,
		DissipationThreshold: 0.25,
		Power:                2,
	}

	model := NewModel()
	N := 16
	field := NewField("conc", N*N, nil)
	model.AddField(field)
	model.RegisterImplicitTerm("SPECTRAL_VISC", &spectral, nil)
	model.AddEquation("dconc/dt = SPECTRAL_VISC")
	model.Init()

	rhs := model.RHS[0]
	if len(rhs.Denum) != 1 {
		t.Errorf("Expected one implicit term got %d\n", len(rhs.Denum))
	}

	if len(rhs.Terms) != 0 {
		t.Errorf("Expected 0 explicit terms, got %d\n", len(rhs.Terms))
	}

	freq := NewFFTW([]int{N, N}).Freq
	res := make([]complex128, N*N)
	rhs.Denum[0](freq, 0.0, res)
	tol := 1e-10
	for i := 0; i < N*N; i++ {
		fVec := freq(i)
		f := Dot(fVec, fVec)
		expect := -spectral.Eps * interpolant(math.Sqrt(f), spectral.DissipationThreshold) * f
		re := real(res[i])
		im := imag(res[i])
		if math.Abs(re-expect) > tol || math.Abs(im) > tol {
			t.Errorf("Expected %f got (%f, %f)\n", expect, re, im)
		}
	}
}
