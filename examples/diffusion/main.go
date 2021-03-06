// +build ignore

package main

import (
	"io/ioutil"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfutil"
)

func main() {
	nx := 128
	ny := 128
	dt := 0.1
	domainSize := []int{nx, ny}
	model := pf.NewModel()
	conc := pf.NewField("conc", nx*ny, nil)
	center := pfutil.NodeIdx(domainSize, []int{nx / 2, ny / 2})

	// Initialize the center
	conc.Data[center] = complex(1.0, 0.0)
	model.AddField(conc)
	model.AddEquation("dconc/dt = LAP conc")

	// Initialize solver
	solver := pf.NewSolver(&model, domainSize, dt)
	model.Summarize()

	// Add a monitor at the center
	monitor := pf.NewPointMonitor(center, "conc")
	solver.AddMonitor(&monitor)

	// Initialize uint8 IO
	out := pf.NewFloat64IO("diffusion2D")
	solver.AddCallback(out.SaveFields)

	// Solve the equation
	solver.Solve(10, 10)

	// Save the monitor
	res := solver.JSONifyMonitors()
	ioutil.WriteFile("diffusionMonitor.json", res, 0644)

	// Write XDMF
	pf.WriteXDMF("diffusion.xdmf", []string{"conc"}, "diffusion2D", 10, domainSize)
}
