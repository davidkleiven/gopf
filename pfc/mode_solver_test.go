package pfc

import (
	"gonum.org/v1/gonum/floats"
	"math"
	"testing"
)

func TestUnitCellIntegrals(t *testing.T) {
	tol := 1e-6
	cellSq := SC2D(4.0)
	recSq2D := cellSq.Reciprocal()
	cellCube := SC3D(4.0)
	recSq3D := cellCube.Reciprocal()
	for i, test := range []struct {
		ModeSolver ModeSolver
		Powers     []int
		Expect     float64
	}{
		{
			ModeSolver: ModeSolver{
				Miller:         []Miller{Miller{H: 1, K: 1}},
				ReciprocalCell: recSq2D,
			},
			Powers: []int{2},
			Expect: 4.0, // num. equiv(=4) * inversion symm (=2) * int cos^2 (=1/2)
		},
		{
			ModeSolver: ModeSolver{
				Miller:         []Miller{Miller{H: 1, K: 1}},
				ReciprocalCell: recSq2D,
			},
			Powers: []int{3},
			Expect: 0.0, // int cos^3 (=0)
		},
		{
			ModeSolver: ModeSolver{
				Miller:         []Miller{Miller{H: 1, K: 1, L: 1}},
				ReciprocalCell: recSq3D,
			},
			Powers: []int{2},
			Expect: 8.0, // num. equiv(=9) * inversion symm (=2) * int cos^2 (=1/2)
		},
		{
			ModeSolver: ModeSolver{
				Miller:         []Miller{Miller{H: 1, K: 1, L: 1}},
				ReciprocalCell: recSq3D,
			},
			Powers: []int{3},
			Expect: 0.0, // int cos^3 (=0)
		},
		{
			ModeSolver: ModeSolver{
				Miller: []Miller{Miller{H: 1, K: 1, L: 1},
					Miller{H: 2, K: 0, L: 0}},
				ReciprocalCell: recSq3D,
			},
			Powers: []int{2, 1},

			// The following integrals gives a non-zero contribution (2*pi skipped inside cosined
			// below). Duplications are obtained as follows
			// A factor 2 for the inversion symmetry of the first term
			// A factor 2 for the inversion symmetry of the second term
			// A factor 3 for cyclic permutations
			// A factor 2 for cross terms when raising the first expression to power 2
			// Each integral listed below occurs 24 times
			// cos(x+y+z)*cos(x-y-z)*cos(2x) = 0.25 // Num. 2*2*3*2 = 24
			// cos(x+y-z)*cos(x-y+z)*cos(2x) = 0.25 // Num. 2*2*3*2 = 24
			// cos(x-y+z)*cos(x+y-z)*cos(2x) = 0.25 // Num. 2*2*3*2 = 24
			// cos(-x+y+z)*cos(x+y+z)*cos(2x) = 0.25 // Num. 2*2*3*2 = 24
			Expect: 96.0 / 4.0,
		},
	} {
		integral := test.ModeSolver.UnitcellIntegral(test.Powers, 16)
		if math.Abs(integral-test.Expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.Expect, integral)
		}
	}
}

func TestSolve(t *testing.T) {
	sc := SC3D(4.05)
	rcell := sc.Reciprocal()
	for _, test := range []struct {
		ModeSolver ModeSolver

		// Expected amplitudes. FCC and BCC are taken from
		// Greenwood, M., Rottler, J., & Provatas, N. (2011).
		// Phase-field-crystal methodology for modeling of structural
		// transformations. Physical review e, 83(3), 031601.
		ExpectAmp []float64
	}{
		{
			ModeSolver: ModeSolver{
				IdealMix: IdealMix{C3: 1.0, C4: 1.0},
				Miller: []Miller{
					Miller{H: 1, K: 1, L: 1},
					Miller{H: 2, K: 0, L: 0},
				},
				ReciprocalCell: rcell,
			},
			ExpectAmp: []float64{0.14274, 0.10648},
		},
		{
			ModeSolver: ModeSolver{
				IdealMix: IdealMix{C3: 1.0, C4: 1.0},
				Miller: []Miller{
					Miller{H: 1, K: 1, L: 0},
					Miller{H: 2, K: 0, L: 0},
				},
				ReciprocalCell: rcell,
			},
			ExpectAmp: []float64{0.13237, 0.05573},
		},
	} {
		// TODO: Currently this test just checks that the solver runs
		// but we should compare the amplitudes being calculated with
		// reference data
		res := test.ModeSolver.Solve(&SinglePointExplorer{X0: []float64{0.1, 0.1}})

		if res.Energy >= 0.0 {
			t.Errorf("Best energy %f. Liquid phase is most stable...", res.Energy)
		}

		if !floats.EqualApprox(res.Amplitudes, test.ExpectAmp, 1e-5) {
			t.Errorf("Expected amplitudes %v\nGot%v\n", test.ExpectAmp, res.Amplitudes)
		}
	}
}

func TestMatchElasticTensor(t *testing.T) {
	cell := SC3D(4.0)
	rec := cell.Reciprocal()
	mode := ModeSolver{
		ReciprocalCell: rec,
		Miller: []Miller{
			Miller{H: 1, K: 1, L: 1},
			Miller{H: 2, K: 0, L: 0},
		},
	}

	B := 60.0
	shear := 40.0
	target := make(map[string]float64)
	target["bulk"] = B
	target["shear"] = shear
	result := mode.MatchElasticProp(target)

	tol := 1e-10
	if math.Abs(result.Fit["bulk"]-B) > tol {
		t.Errorf("Expected %f. Got %f\n", B, result.Fit["bulk"])
	}

	if math.Abs(result.Fit["shear"]-shear) > tol {
		t.Errorf("Expected %f. Got %f\n", shear, result.Fit["shear"])
	}
}
