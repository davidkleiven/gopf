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
