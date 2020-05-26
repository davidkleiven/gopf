package pfutil

import (
	"math"
	"testing"
)

func TestIndex(t *testing.T) {
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
		idx := test.Grid.Index(test.Pos)
		if idx != test.Expect {
			t.Errorf("Test #%d: Expected %d got %d\n", i, test.Expect, idx)
		}
		pos := test.Grid.Pos(idx)
		for j, p := range test.Pos {
			if p != pos[j] {
				t.Errorf("Test #%d:\nExpected\n%v\nGot\n%v\n", i, test.Pos, pos)
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

func TestFromComplex(t *testing.T) {
	grid := NewGrid([]int{2, 2})
	carray := []complex128{
		complex(1.0, 0.0), complex(2.0, 0.0),
		complex(3.0, 0.0), complex(4.0, 0.0),
	}
	grid.FromComplex(carray)
	expect := []float64{1.0, 2.0, 3.0, 4.0}
	tol := 1e-10
	for i := range carray {
		if math.Abs(grid.Data[i]-expect[i]) > tol {
			t.Errorf("Expected %f got %f\n", expect[i], grid.Data[i])
		}
	}
}
func TestRotate2D(t *testing.T) {
	grid := NewGrid([]int{8, 8})
	grid.Set([]int{5, 4}, 1.0)
	grid.Rotate2D(math.Pi / 2.0)
	expectPos := []int{4, 3}
	if math.Abs(grid.Get(expectPos)-1.0) > 1e-10 {
		t.Errorf("Expected 1.0 got %v\n", expectPos)
	}
}

func TestCopy(t *testing.T) {
	grid := NewGrid([]int{4, 4})
	for i := range grid.Data {
		grid.Data[i] = float64(i)
	}
	gridCpy := grid.Copy()
	for i := range grid.Dims {
		if grid.Dims[i] != gridCpy.Dims[i] {
			t.Errorf("Expected %d got %d\n", grid.Dims[i], gridCpy.Dims[i])
		}
	}

	for i := range grid.Data {
		if math.Abs(grid.Data[i]-gridCpy.Data[i]) > 1e-10 {
			t.Errorf("Expected %f got %f\n", grid.Data[i], gridCpy.Data[i])
		}
	}
}
