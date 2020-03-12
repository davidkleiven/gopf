package pf

import (
	"math"
	"testing"
)

func TestPairCorrelation(t *testing.T) {
	dk := 0.1
	for i, test := range []struct {
		Corr           ReciprocalSpacePairCorrelation
		ExpectedPeaks  []float64
		ExpectedMinima []float64
	}{
		{
			Corr: ReciprocalSpacePairCorrelation{
				Peaks: SquareLattice2D(1.0, 1.0),
			},
			ExpectedPeaks:  []float64{2.0 * math.Pi, 2.0 * math.Pi * math.Sqrt(2.0)},
			ExpectedMinima: []float64{math.Pi * (1.0 + math.Sqrt(2.0))},
		},
		{
			Corr: ReciprocalSpacePairCorrelation{
				Peaks: TriangularLattice2D(1.0, 1.0),
			},
			ExpectedPeaks:  []float64{2.0 * math.Pi},
			ExpectedMinima: []float64{},
		},
	} {
		for _, expPeak := range test.ExpectedPeaks {
			peak := test.Corr.Eval(expPeak)
			peakPluss := test.Corr.Eval(expPeak + dk)
			peakMinus := test.Corr.Eval(expPeak - dk)

			if peak < peakPluss || peak < peakMinus {
				t.Errorf("Test #%d: Expected %f < %f && %f < %f", i, peakMinus, peak, peakPluss, peak)
			}
		}

		for _, expMin := range test.ExpectedMinima {
			minimum := test.Corr.Eval(expMin)
			minPluss := test.Corr.Eval(expMin + dk)
			minMinus := test.Corr.Eval(expMin - dk)

			if minimum > minMinus || minimum > minPluss {
				t.Errorf("Test #%d: minumum: Expected %f > %f and %f > %f", i, minPluss, minimum, minMinus, minimum)
			}
		}
	}
}

func TestPairCorrelationTerm(t *testing.T) {
	pair := PairCorrlationTerm{
		PairCorrFunc: ReciprocalSpacePairCorrelation{
			EffTemp: 0.0,
			Peaks: []Peak{
				Peak{
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
