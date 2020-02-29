package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

func TestExponentialDecay(t *testing.T) {
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
	model.AddEquation("dfield/dt = rate*field")
	model.Init()
	model.Summarize()

	stepper := Euler{
		Dt: 0.001,
		FT: sfft.NewFFT2(N, N),
	}
	nsteps := 1000
	finalTime := float64(nsteps) * stepper.Dt
	expect := math.Exp(-finalTime)
	stepper.Propagate(nsteps, &model)
	tol := 1e-3
	for i := range field.Data {
		re := real(field.Data[i])
		im := imag(field.Data[i])

		if math.Abs(re-expect) > tol || math.Abs(im) > tol {
			t.Errorf("Node %d: Expected %f got (%f, %f)\n", i, expect, re, im)
		}
	}

	expectTime := float64(nsteps) * stepper.Dt
	if math.Abs(stepper.GetTime()-expectTime) > 1e-10 {
		t.Errorf("Expected time: %f. Got %f", expectTime, stepper.GetTime())
	}
}

func TestEulerSquareDecary(t *testing.T) {
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

	stepper := Euler{
		Dt: 0.001,
		FT: sfft.NewFFT2(N, N),
	}
	nsteps := 1000
	finalTime := float64(nsteps) * stepper.Dt
	expect := AnalyticalSimpleModel(finalTime)
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
