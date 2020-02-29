package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

func AnalyticalSimpleModel(t float64) float64 {
	return 1.0 / (1.0 + t)
}

func TestSimpleModel(t *testing.T) {
	N := 8
	field := NewField("field", N*N, nil)
	for i := range field.Data {
		field.Data[i] = complex(1.0, 0.0)
	}

	model := NewModel()
	model.AddField(field)
	model.AddScalar(Scalar{
		Name:  "rate",
		Value: complex(-1.0, 0.0),
	})

	model.AddEquation("dfield/dt = rate*field^2")
	model.Init()
	model.Summarize()

	stepper := RK4{
		Dt: 0.1,
		FT: sfft.NewFFT2(N, N),
	}

	nsteps := 10
	stepper.Propagate(nsteps, &model)
	finalTime := float64(nsteps) * stepper.Dt

	expect := AnalyticalSimpleModel(finalTime)
	tol := 1e-6
	for i, v := range field.Data {
		re := real(v)
		im := imag(v)
		if math.Abs(re-expect) > tol || math.Abs(im) > tol {
			t.Errorf("Node: %d: Expected %f, got (%f, %f)\n", i, expect, re, im)
		}
	}

	if math.Abs(finalTime-stepper.GetTime()) > 1e-10 {
		t.Errorf("Expected time: %f got %f\n", finalTime, stepper.GetTime())
	}
}

func AnalyticalImplicit(t float64, c0 float64) float64 {
	A := 1.0/c0 - 1.0
	return math.Exp(t) / (A + math.Exp(t))
}

func TestWithImplicit(t *testing.T) {
	N := 8
	field := NewField("field", N*N, nil)
	c0 := 0.5
	for i := range field.Data {
		field.Data[i] = complex(c0, 0.0)
	}

	model := NewModel()
	model.AddField(field)
	model.AddScalar(Scalar{
		Name:  "rate",
		Value: complex(-1.0, 0.0),
	})

	model.AddEquation("dfield/dt = field + rate*field^2")
	model.Init()
	model.Summarize()

	stepper := RK4{
		Dt: 0.01,
		FT: sfft.NewFFT2(N, N),
	}

	nsteps := 100
	finalTime := float64(nsteps) * stepper.Dt
	expect := AnalyticalImplicit(finalTime, c0)

	stepper.Propagate(nsteps, &model)
	tol := 1e-3
	for i := range field.Data {
		re := real(field.Data[i])
		im := imag(field.Data[i])

		if math.Abs(re-expect) > tol || math.Abs(im) > tol {
			t.Errorf("Node %d: Expected %f got (%f, %f)\n", i, expect, re, im)
		}
	}

}
