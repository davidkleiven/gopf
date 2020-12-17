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

	model.AddEquation("ddensity/dt = -NEG_VAL_PENALTY")

	// Evaluate the rhs
	model.Init()
	
	solver := NewSolver(&model, []int{2, 2}, 1.0)
	solver.Solve(1, 1)

	// Solver should perform one step
	expect := []float64{-1.0 - nvp.Penalty(-1.0), 0.0, 0.0, 0.0}

	tol := 1e-6
	for i := range field.Data {
		re, im := real(field.Data[i]), imag(field.Data[i])
		if math.Abs(im) > tol || math.Abs(re - expect[i]) > tol {
			t.Errorf("Expected\n%v\nGot\n%v\n", expect, field.Data)
		}
	}
}

func TestConservedNegativeValuePenalty(t *testing.T) {
	// Make sure that the function can be registered
	model := NewModel()
	nvp := NegativeValuePenalty{
		Prefactor: 1.0,
		Exponent: 3,
		Field: "density",
	}
	
	field := NewField("density", 16, nil)
	integral := 0.0
	for i := range field.Data {
		field.Data[i] = complex(float64(i) - 8.0, 0.0)
		integral += real(field.Data[i])
	}
	model.AddField(field)
	model.RegisterFunction("NEG_VAL_PENALTY", nvp.Evaluate)

	model.AddEquation("ddensity/dt = LAP*NEG_VAL_PENALTY")

	// Evaluate the rhs
	model.Init()
	
	solver := NewSolver(&model, []int{4, 4}, 0.005)
	solver.Solve(2, 100)

	integralAfter := 0.0
	for i := range field.Data {
		integralAfter += real(field.Data[i])
	}

	if math.IsNaN(integralAfter) {
		t.Errorf("Field diverged to infinity")
	}

	if math.Abs(integral - integralAfter) > 1e-6 {
		t.Errorf("Field not conserved when using conserved dynamics. Integral before %f and after %f\n", integral, integralAfter)
	}
}