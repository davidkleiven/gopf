package pfc

import (
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/gonum/mat"
)

func TestCornerLocation(t *testing.T) {
	for i, test := range []struct {
		Cell       Cell
		DomainSize []int
		Expect     *mat.Dense
	}{
		{
			Cell:       SC2D(4.0),
			DomainSize: []int{16.0, 16.0},
			Expect: mat.NewDense(2, 4, []float64{
				0.0, 0.0, 4.0, 4.0,
				0.0, 4.0, 0.0, 4.0,
			}),
		},
		{
			Cell:       SC3D(4.0),
			DomainSize: []int{16.0, 16.0, 16.0},
			Expect: mat.NewDense(3, 8, []float64{
				0.0, 0.0, 0.0, 0.0, 4.0, 4.0, 4.0, 4.0,
				0.0, 0.0, 4.0, 4.0, 0.0, 0.0, 4.0, 4.0,
				0.0, 4.0, 0.0, 4.0, 0.0, 4.0, 0.0, 4.0,
			}),
		},
	} {
		loc := CornersScaledCrd(test.Cell, test.DomainSize)
		if !mat.EqualApprox(loc, test.Expect, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, mat.Formatted(test.Expect), mat.Formatted(loc))
		}
	}
}

func TestBuildCrystal(t *testing.T) {
	kernel := CircleKernel{
		Radius: 8.0,
	}

	for i, test := range []struct {
		Ucell UnitCell
		Grid  pfutil.Grid
		Want  int
	}{
		{
			Ucell: UnitCell{
				Cell:  SC2D(64.0),
				Basis: mat.NewDense(2, 1, []float64{0.0, 0.0}),
			},
			Grid: pfutil.NewGrid([]int{128, 128}),
			Want: 4,
		},
		{
			Ucell: UnitCell{
				Cell: SC2D(64.0),
				Basis: mat.NewDense(2, 2, []float64{0.0, 0.5,
					0.0, 0.5}),
			},
			Grid: pfutil.NewGrid([]int{128, 128}),
			Want: 8,
		},
		{
			Ucell: UnitCell{
				Cell:  Triangular2D(64.0),
				Basis: mat.NewDense(2, 1, []float64{0.0, 0.0}),
			},
			Grid: pfutil.NewGrid([]int{128, 128}),
			Want: 8,
		},
	} {
		BuildCrystal(test.Ucell, &kernel, &test.Grid)

		integral := 0.0
		for j := range test.Grid.Data {
			integral += test.Grid.Data[j]
		}
		numAtoms := int(integral / kernel.Area())

		if numAtoms != test.Want {
			t.Errorf("Test #%d: Expected %d got %d\n", i, test.Want, numAtoms)
		}
	}
}
