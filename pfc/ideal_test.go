package pfc

import (
	"math"
	"testing"
)

func TestIdeal(t *testing.T) {
	tol := 1e-10
	for i, test := range []struct {
		IdealMix IdealMix
		n        float64
		expect   float64
	}{
		{
			IdealMix: IdealMix{
				C3: 0.0,
				C4: 0.0,
			},
			n:      0.5,
			expect: 1.0 / 8.0,
		},
		{
			IdealMix: IdealMix{
				C3: 1.0,
				C4: 1.0,
			},
			n:      0.5,
			expect: 0.109375,
		},
	} {
		res := test.IdealMix.Eval(test.n)
		if math.Abs(res-test.expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.expect, res)
		}
	}
}
