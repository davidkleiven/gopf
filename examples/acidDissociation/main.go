package main

import (
	"github.com/davidkleiven/gopf/pf"
)

func main() {
	nx := 128
	ny := 128
	N := nx * ny
	model := pf.NewModel()
	//K := 1.75e-5
	dt := 0.00001
	rateForward := pf.NewScalar("rateForward", complex(1.0, 0.0))
	rateBackward := pf.NewScalar("rateBackward", complex(1.0, 0.0))
	m1 := pf.NewScalar("m1", complex(-1.0, 0.0))
	model.AddScalar(rateBackward)
	model.AddScalar(rateForward)
	model.AddScalar(m1)

	// Define fields
	ch3cooh := pf.NewField("ch3cooh", N, nil)
	h2o := pf.NewField("h2o", N, nil)

	// Initialize half the space with ch3cooh and the other half with h2o
	for i := 0; i < nx*ny; i++ {
		if i < nx*ny/2 {
			ch3cooh.Data[i] = complex(1.0, 0.0)
		} else {
			h2o.Data[i] = complex(1.0, 0.0)
		}
	}
	h3oPlus := pf.NewField("h3oPlus", N, nil)
	ch3cooMinus := pf.NewField("ch3cooMinus", N, nil)

	// Add fields to the model
	model.AddField(ch3cooh)
	model.AddField(h2o)
	model.AddField(h3oPlus)
	model.AddField(ch3cooMinus)

	// For convenience, define the source terms for products and reactants
	reactantsSource := "rateBackward*h3oPlus*ch3cooMinus + m1*rateForward*ch3cooh*h2o"
	productsSource := "m1*rateBackward*h3oPlus*ch3cooMinus + rateForward*ch3cooh*h2o"

	// Add equations
	model.AddEquation("dch3cooh/dt = LAP ch3cooh + " + reactantsSource)
	model.AddEquation("dh2o/dt = LAP h2o + " + reactantsSource)
	model.AddEquation("dh3oPlus/dt = LAP h3oPlus + " + productsSource)
	model.AddEquation("dch3cooMinus/dt = LAP ch3cooMinus + " + productsSource)

	// Initialize the solver
	domainSize := []int{nx, ny}
	solver := pf.NewSolver(&model, domainSize, dt)

	// Initialize uint8 IO
	out := pf.NewFloat64IO("acidDiss")
	solver.AddCallback(out.SaveFields)

	// Solve the equation
	solver.Solve(10, 10)
}
