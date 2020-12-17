package pf

import (
	"testing"
	"math"
)

func TestDefaultFunction(t *testing.T) {
	nvp := NewDefaultNegativeValuePenalty("myfield")
	tol := 1e-6
	for i, test := range []struct {
		Value float64
		Expect float64
	}{
		{
			Value: 1.0,
			Expect: 0.0,
		},
		{
			Value: 0.0,
			Expect: 0.0,
		},
		{
			Value: -1.0,
			Expect: -2.0*nvp.Prefactor*float64(nvp.Exponent),
		},
	}{
		res := nvp.Penalty(test.Value)
		if math.Abs(res - test.Expect) > tol {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.Expect, res)
		}
	}
}

func TestIsRegisterable(t *testing.T) {
	// Make sure that the function can be registered
	model := NewModel()
	nvp := NewDefaultNegativeValuePenalty("density")
	
	field := NewField("density", 4, nil)
	field.Data[0] = complex(-1.0, 0.0)
	model.AddField(field)
	model.RegisterFunction("NEG_VAL_PENALTY", nvp.Evaluate)

	model.AddEquation("ddensity/dt = NEG_VAL_PENALTY")

	// Evaluate the rhs
	model.Init()
	
	solver := NewSolver(&model, []int{2, 2}, 1.0)
	solver.Solve(1, 1)

	// Solver should perform one step
	expect := []float64{-1.0 + nvp.Penalty(-1.0), 0.0, 0.0, 0.0}

	tol := 1e-6
	for i := range field.Data {
		re, im := real(field.Data[i]), imag(field.Data[i])
		if math.Abs(im) > tol || math.Abs(re - expect[i]) > tol {
			t.Errorf("Expected\n%v\nGot\n%v\n", expect, field.Data)
		}
	}
}