package pf

import (
	"math"
	"testing"
)

func TestVandeven(t *testing.T) {
	type XY struct {
		x float64
		y float64
	}
	for i, test := range []struct {
		expect []XY
		order  int
	}{
		{
			expect: []XY{XY{x: 0.0, y: 1.0}, XY{x: 0.5, y: 0.5}, XY{x: 1.0, y: 0.0}},
			order:  3,
		},
		{
			expect: []XY{XY{x: 0.0, y: 1.0}, XY{x: 0.5, y: 0.5}, XY{x: 1.0, y: 0.0}},
			order:  5,
		},
		{
			expect: []XY{XY{x: 0.0, y: 1.0}, XY{x: 0.5, y: 0.5}, XY{x: 1.0, y: 0.0}},
			order:  10,
		},
	} {
		filter := NewVandeven(test.order)
		tol := 1e-4
		for _, xy := range test.expect {
			value := filter.Eval(xy.x)
			if math.Abs(xy.y-value) > tol {
				t.Errorf("Test %d: Order %d Expected %f got %f\n", i, test.order, xy.y, value)
			}
		}
	}
}
