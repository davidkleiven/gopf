package pfc

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestVolumes(t *testing.T) {
	tol := 1e-10
	for i, test := range []struct {
		Cell   Cell
		Volume float64
	}{
		{
			Cell:   SC2D(4.0),
			Volume: 16.0,
		},
		{
			Cell:   SC3D(4.0),
			Volume: 64.0,
		},
		{
			Cell:   FCC(4.0),
			Volume: 16.0,
		},
		{
			Cell:   BCC(4.0),
			Volume: 32.0,
		},
		{
			Cell:   Triangular2D(4.0),
			Volume: 8.0,
		},
	} {
		vol := test.Cell.Volume()
		if math.Abs(vol-test.Volume) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.Volume, vol)
		}
	}
}

func TestReciprocal(t *testing.T) {
	tol := 1e-10
	for i, test := range []struct {
		Cell   Cell
		Expect *mat.Dense
	}{
		{
			Cell:   SC2D(4.0),
			Expect: mat.NewDense(2, 2, []float64{0.25, 0.0, 0.0, 0.25}),
		},
		{
			Cell:   SC3D(4.0),
			Expect: mat.NewDense(3, 3, []float64{0.25, 0.0, 0.0, 0.0, 0.25, 0.0, 0.0, 0.0, 0.25}),
		},
		{
			Cell:   Triangular2D(4.0),
			Expect: mat.NewDense(2, 2, []float64{0.25, 0.0, -0.25, 0.5}),
		},
	} {
		res := test.Cell.Reciprocal()
		r, c := test.Cell.CellVec.Dims()
		for j := 0; j < r; j++ {
			for k := 0; k < c; k++ {
				expect := 2.0 * math.Pi * test.Expect.At(j, k)
				if math.Abs(res.CellVec.At(j, k)-expect) > tol {
					t.Errorf("Test #%d: Expected %f got %f\n", i, expect, res.CellVec.At(j, k))
				}
			}
		}
	}
}

func TestHKLVector(t *testing.T) {
	tol := 1e-10
	cell := SC3D(4.0)
	rec := cell.Reciprocal()
	miller := Miller{1, 1, 1}
	g := rec.HKLVector(miller)
	length := math.Sqrt(g[0]*g[0] + g[1]*g[1] + g[2]*g[2])
	expect := 2.0 * math.Pi * math.Sqrt(3.0) / 4.0
	if math.Abs(expect-length) > tol {
		t.Errorf("Expected length of hkl vector %f got %f\n", expect, length)
	}
}

func TestChangeHKLVector(t *testing.T) {
	a := 4.0
	delta := 0.01
	cell := SC3D(a)
	rec := cell.Reciprocal()
	change := mat.NewDense(3, 3, []float64{delta * a, 0.0, 0.0,
		0.0, delta * a, 0.0,
		0.0, 0.0, delta * a})

	// The real-space cell is expanded
	expectedChange := -2.0 * math.Pi * delta / a
	tol := 1e-10
	for i, miller := range []Miller{
		Miller{H: 1, K: 0, L: 0},
		Miller{H: 0, K: 1, L: 0},
		Miller{H: 0, K: 0, L: 1},
	} {
		dk := rec.ChangeHKLVector(miller, change)
		value := dk[0]
		if miller.K == 1 {
			value = dk[1]
		} else if miller.L == 1 {
			value = dk[2]
		}

		if math.Abs(value-expectedChange) > tol {
			t.Errorf("Test #%d: Expected change %f. Got %f\n", i, expectedChange, value)
		}
	}
}

func TestCubicDensity(t *testing.T) {
	for i, test := range []struct {
		Miller Miller
		Expect float64
	}{
		{
			Miller: Miller{H: 1, K: 0, L: 0},
			Expect: 1.0,
		},
		{
			Miller: Miller{H: 0, K: 0, L: 1},
			Expect: 1.0,
		},
		{
			Miller: Miller{H: 1, K: 1, L: 0},
			Expect: 1.0 / math.Sqrt(2.0),
		},
		{
			Miller: Miller{H: 0, K: 1, L: 1},
			Expect: 1.0 / math.Sqrt(2.0),
		},
		{
			Miller: Miller{H: 1, K: 1, L: 1},
			Expect: 1.0 / math.Sqrt(3.0),
		},
	} {
		density := CubicUnitCellDensity(test.Miller)
		tol := 1e-10
		if math.Abs(density-test.Expect) > tol {
			t.Errorf("Test #%d: Expected %f, got %f\n", i, test.Expect, density)
		}
	}
}
