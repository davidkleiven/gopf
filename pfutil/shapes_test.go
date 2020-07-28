package pfutil

import (
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestShapes(t *testing.T) {
	for i, test := range []struct {
		shape Shape
		grid  Grid
		want  []float64
	}{
		{
			shape: &Box{Diagonal: []float64{5.1, 5.1}},
			grid:  NewGrid([]int{8, 8}),
			want: []float64{1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		},
		{
			shape: &Ball{Radius: 2.1},
			grid:  NewGrid([]int{8, 8}),
			want: []float64{0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0,
				0.0, 1.0, 1.0, 1.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
		},
	} {
		trans := Translation([]float64{-2.0, -2.0})
		Draw(test.shape, &test.grid, &trans, 1.0)

		if !floats.EqualApprox(test.want, test.grid.Data, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, test.want, test.grid.Data)
		}
	}
}
