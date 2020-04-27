package pf

import "testing"
import "math"

func TestBoundsTransform(t *testing.T) {
	bounds := BoundTransform{
		Min: -2.0,
		Max: 2.0,
	}
	for i, test := range []struct {
		x      float64
		expect float64
	}{
		{
			x:      0.0,
			expect: 0.0,
		},
		{
			x:      1.0,
			expect: 1.0,
		},
	} {
		tol := 1e-10
		y := bounds.Forward(test.x)
		if math.Abs(y-test.expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.expect, y)
		}

		xBack := bounds.Backward(y)
		if math.Abs(xBack-test.x) > tol {
			t.Errorf("Test #%d: Backward ended with %f, expectecd %f\n", i, xBack, test.x)
		}
	}
}
