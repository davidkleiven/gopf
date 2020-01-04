package main

import (
	"io/ioutil"

	"github.com/davidkleiven/gopf/pf"
)

func main() {
	nx := 128
	ny := 128
	dt := 0.1
	domainSize := []int{nx, ny}
	model := pf.NewModel()
	conc := pf.NewField("conc", nx*ny, nil)
	center := pf.NodeIdx(domainSize, []int{nx / 2, ny / 2})

	// Initialize the center
	conc.Data[center] = complex(1.0, 0.0)
	model.AddField(conc)
	model.AddEquation("dconc/dt = LAP conc")

	// Initialize solver
	solver := pf.NewSolver(&model, domainSize, dt)

	// Add a monitor at the center
	monitor := pf.NewPointMonitor(center, "conc")
	solver.AddMonitor(monitor)

	// Initialize uint8 IO
	out := pf.NewUint8IO("diffusion2D")
	solver.AddCallback(out.SaveFields)

	// Solve the equation
	solver.Solve(10, 10)

	// Save the monitor
	res := solver.JSONifyMonitors()
	ioutil.WriteFile("diffusionMonitor.json", res, 0644)
}
