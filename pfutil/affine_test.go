package pfutil

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestAffine(t *testing.T) {
	for i, test := range []struct {
		transformation Affine
		vec            []float64
		want           []float64
	}{
		{
			transformation: Identity(),
			vec:            []float64{0.4, 0.5, 0.2},
			want:           []float64{0.4, 0.5, 0.2},
		},
		{
			transformation: Identity(),
			vec:            []float64{0.4, 0.5},
			want:           []float64{0.4, 0.5},
		},
		{
			transformation: Translation([]float64{2.0, -1.0, 2.0}),
			vec:            []float64{1.0, 1.0, 1.0},
			want:           []float64{3.0, 0.0, 3.0},
		},
		{
			transformation: Translation([]float64{2.0, -1.0}),
			vec:            []float64{1.0, 1.0},
			want:           []float64{3.0, 0.0},
		},
		{
			transformation: Scaling([]float64{0.5, 2.0, 3.0}),
			vec:            []float64{2.0, 4.0, 6.0},
			want:           []float64{1.0, 8.0, 18.0},
		},
		{
			transformation: Scaling([]float64{0.5, 2.0}),
			vec:            []float64{2.0, 4.0},
			want:           []float64{1.0, 8.0},
		},
		{
			transformation: RotZ(math.Pi / 4.0),
			vec:            []float64{1.0, 1.0, 4.0},
			want:           []float64{0.0, math.Sqrt(2.0), 4.0},
		},
		{
			transformation: RotZ(math.Pi / 4.0),
			vec:            []float64{1.0, 1.0},
			want:           []float64{0.0, math.Sqrt(2.0)},
		},
		{
			transformation: RotY(math.Pi / 4.0),
			vec:            []float64{1.0, 4.0, 1.0},
			want:           []float64{math.Sqrt(2.0), 4.0, 0.0},
		},
		{
			transformation: RotX(math.Pi / 4.0),
			vec:            []float64{4.0, 1.0, 1.0},
			want:           []float64{4.0, 0.0, math.Sqrt(2.0)},
		},
	} {
		test.transformation.Apply(test.vec)
		if !floats.EqualApprox(test.vec, test.want, 1e-10) {
			t.Errorf("Test #%d: Expected %v got %v\n", i, test.want, test.vec)
		}
	}
}

func TestAffineAppend(t *testing.T) {
	for i, test := range []struct {
		transformations []Affine
		vec             []float64
		want            []float64
	}{
		{
			transformations: []Affine{Translation([]float64{1.0, 1.0}), RotZ(math.Pi / 4.0)},
			vec:             []float64{0.0, 0.0},
			want:            []float64{0.0, math.Sqrt(2.0)},
		},
		{
			transformations: []Affine{RotZ(math.Pi / 4.0), Translation([]float64{1.0, 1.0})},
			vec:             []float64{math.Sqrt(2.0), 0.0},
			want:            []float64{2.0, 2.0},
		},
		{
			transformations: []Affine{Scaling([]float64{0.5, 0.5}), Translation([]float64{1.0, 1.0})},
			vec:             []float64{1.0, 1.0},
			want:            []float64{1.5, 1.5},
		},
		{
			transformations: []Affine{Translation([]float64{1.0, 1.0}), Scaling([]float64{0.5, 0.5})},
			vec:             []float64{1.0, 1.0},
			want:            []float64{1.0, 1.0},
		},
	} {
		trans := test.transformations[0]
		for i := 1; i < len(test.transformations); i++ {
			trans.Append(test.transformations[i])
		}
		trans.Apply(test.vec)

		if !floats.EqualApprox(test.vec, test.want, 1e-10) {
			t.Errorf("Test #%d: Expected %v got %v\n", i, test.want, test.vec)
		}
	}
}
