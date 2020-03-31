package pf

import (
	"github.com/davidkleiven/gopf/pfc"
	"math"
	"testing"
)

func TestPairCorrelationTerm(t *testing.T) {
	pair := PairCorrlationTerm{
		PairCorrFunc: pfc.ReciprocalSpacePairCorrelation{
			EffTemp: 0.0,
			Peaks: []pfc.Peak{
				pfc.Peak{
					PlaneDensity: 1,
					Location:     1.0,
					Width:        100.0,
					NumPlanes:    1,
				},
			},
		},
		Field:     "myfield",
		Prefactor: 1.0,
	}

	N := 16
	field := NewField("myfield", N*N, nil)

	// Insert fourier transformed fields
	for i := range field.Data {
		field.Data[i] = complex(0.1*float64(i), 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["myfield"] = field
	function := pair.Construct(bricks)
	res := make([]complex128, N*N)

	freq := func(i int) []float64 {
		return []float64{float64(i), float64(2 * i)}
	}
	function(freq, 0.0, res)

	for i := range field.Data {
		f := freq(i)
		fRad := 2.0 * math.Pi * math.Sqrt(Dot(f, f))
		wSq := math.Pow(pair.PairCorrFunc.Peaks[0].Width, 2)
		factor := math.Exp(-0.5 * (fRad - 1.0) * (fRad - 1.0) / wSq)
		expect := factor * real(field.Data[i])
		re := real(res[i])

		if math.Abs(re-expect) > 1e-10 {
			t.Errorf("Expected %f got %f\n", expect, re)
		}
	}
}
