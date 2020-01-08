package elasticity

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestIsotropic(t *testing.T) {
	B := 61.4
	poisson := 0.3
	tensor := Isotropic(B, poisson)
	data := make([]float64, 81)
	copy(data, tensor.Data)
	for i, test := range []struct {
		axis  int
		angle float64
	}{
		{
			axis:  0,
			angle: 14.0 * math.Pi / 180.0,
		},
		{
			axis:  1,
			angle: 56.0 * math.Pi / 180.0,
		},
		{
			axis:  2,
			angle: -56.0 * math.Pi / 180.0,
		},
	} {
		rot := RotationMatrix(test.angle, test.axis)
		tensor.Rotate(rot)
		if !floats.EqualApprox(tensor.Data, data, 1e-10) {
			t.Errorf("Test #%d: Isotropic tensor is not invariant under rotations", i)
		}
	}
}
