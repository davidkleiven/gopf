package pfc

import (
	"math"
	"testing"
)

func TestQuadrature2D(t *testing.T) {
	tol := 1e-12
	for i, test := range []struct {
		f      func(float64, float64) float64
		expect float64
		n      int
	}{
		{
			f:      func(x float64, y float64) float64 { return 1.0 },
			expect: 1.0,
			n:      1,
		},
		{
			f:      func(x float64, y float64) float64 { return 1.0 },
			expect: 1.0,
			n:      3,
		},
		{
			f:      func(x float64, y float64) float64 { return x*x},
			expect: 1./3.,
			n:      3,
		},
		{
			f:      func(x float64, y float64) float64 { return x*x*y*y},
			expect: 1./9.,
			n:      3,
		},
	} {
		res := QuadSquare(test.f, test.n)
		if math.Abs(res-test.expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.expect, res)
		}
	}
}

func TestQuadrature3D(t *testing.T) {
	tol := 1e-12
	for i, test := range []struct {
		f      func(float64, float64, float64) float64
		expect float64
		n      int
	}{
		{
			f:      func(x float64, y float64, z float64) float64 { return 1.0 },
			expect: 1.0,
			n:      1,
		},
		{
			f:      func(x float64, y float64, z float64) float64 { return 1.0 },
			expect: 1.0,
			n:      3,
		},
		{
			f:      func(x float64, y float64, z float64) float64 { return x*x},
			expect: 1./3.,
			n:      3,
		},
		{
			f:      func(x float64, y float64, z float64) float64 { return x*x*y*y},
			expect: 1./9.,
			n:      3,
		},
		{
			f:      func(x float64, y float64, z float64) float64 { return x*x*y*y*z},
			expect: 1./18.,
			n:      3,
		},
	} {
		res := QuadCube(test.f, test.n)
		if math.Abs(res-test.expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.expect, res)
		}
	}
}