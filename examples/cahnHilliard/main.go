// +build ignore

package main

import (
	"math/rand"

	"github.com/davidkleiven/gopf/pf"
)

func main() {
	nx := 128
	ny := 128
	dt := 0.1
	domainSize := []int{nx, ny}
	model := pf.NewModel()
	conc := pf.NewField("conc", nx*ny, nil)

	// Initialize with random concentration
	r := rand.New(rand.NewSource(0))
	for i := range conc.Data {
		conc.Data[i] = complex(2.0*r.Float64()-1.0, 0.0)
	}

	// Add constants
	gamma := pf.NewScalar("gamma", complex(2.0, 0.0)) // Gradient coefficient
	m1 := pf.NewScalar("m1", complex(-1.0, 0.0))      // -1.0
	model.AddScalar(gamma)
	model.AddScalar(m1)

	// Initialize the center
	model.AddField(conc)
	model.AddEquation("dconc/dt = LAP conc^3 + m1*LAP conc + m1*gamma*LAP^2 conc")

	// Initialize solver
	solver := pf.NewSolver(&model, domainSize, dt)
	model.Summarize()

	// Initialize uint8 IO
	out := pf.NewUint8IO("cahnHilliard2D")
	solver.AddCallback(out.SaveFields)

	// Solve the equation
	solver.Solve(10, 10)
}
