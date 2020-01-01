package pf

import (
	"math"
	"testing"
)

func TestLaplacianN(t *testing.T) {
	N := 4
	data := make([]complex128, N)
	for i := range data {
		data[i] = complex(float64(i), 0.0)
	}

	Freq := func(i int) []float64 {
		return []float64{float64(i)}
	}

	for i, test := range []struct {
		power  int
		expect func() []complex128
	}{
		{
			power: 1,
			expect: func() []complex128 {
				res := make([]complex128, N)
				for j := range res {
					res[j] = complex(-2.0*math.Pi*2.0*math.Pi*float64(j)*float64(j), 0.0) * data[j]
				}
				return res
			},
		},
		{
			power: 2,
			expect: func() []complex128 {
				res := make([]complex128, N)
				for j := range res {
					res[j] = complex(math.Pow(2.0*math.Pi, 4.0)*float64(j*j*j*j), 0.0) * data[j]
				}
				return res
			},
		},
	} {
		tmpData := make([]complex128, N)
		copy(tmpData, data)
		var lap LaplacianN
		lap.Power = test.power
		lap.Eval(Freq, tmpData)

		expect := test.expect()
		if !CmplxEqualApprox(expect, tmpData, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, expect, tmpData)
		}
	}
}
