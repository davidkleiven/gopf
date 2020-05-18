package pf

import (
	"math"
	"testing"
)

func TestIndex(t *testing.T) {
	work := make([]int, 3)
	for i, test := range []struct {
		Grid   Grid
		Pos    []int
		Expect int
	}{
		{
			Grid:   NewGrid([]int{3, 4}),
			Pos:    []int{1, 2},
			Expect: 6,
		},
		{
			Grid:   NewGrid([]int{3, 4, 2}),
			Pos:    []int{2, 3, 1},
			Expect: 12 + 2*4 + 3,
		},
	} {
		idx := test.Grid.index(test.Pos)
		if idx != test.Expect {
			t.Errorf("Test #%d: Expected %d got %d\n", i, test.Expect, idx)
		}
		test.Grid.Pos(idx, work)
		for j, p := range test.Pos {
			if p != work[j] {
				t.Errorf("Test #%d:\nExpected\n%v\nGot\n%v\n", i, test.Pos, work)
				break
			}
		}
	}
}

func TestGetSet(t *testing.T) {
	grid := NewGrid([]int{2, 5, 7})
	grid.Set([]int{2, 3, 4}, 5.0)
	value := grid.Get([]int{2, 3, 4})
	if math.Abs(value-5.0) > 1e-10 {
		t.Errorf("Expected 5.0, got %f\n", value)
	}
}

func TestToComplex(t *testing.T) {
	grid := NewGrid([]int{2, 2})
	grid.Data = []float64{1.0, 2.0, 3.0, 4.0}
	carray := grid.ToComplex()
	expect := []complex128{
		complex(1.0, 0.0), complex(2.0, 0.0),
		complex(3.0, 0.0), complex(4.0, 0.0),
	}

	tol := 1e-10
	for i := range carray {
		re := real(carray[i])
		im := imag(carray[i])
		reExp := real(expect[i])
		imExp := imag(expect[i])

		if math.Abs(re-reExp) > tol || math.Abs(im-imExp) > tol {
			t.Errorf("Expected (%f, %f) got (%f, %f)\n", reExp, imExp, re, im)
		}
	}
}
