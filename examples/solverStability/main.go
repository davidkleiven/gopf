package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/davidkleiven/gopf/pf"
)

// DerivChemPot is the derivative of the chemical potential
func DerivChemPot(i int, bricks map[string]pf.Brick) complex128 {
	x := bricks["conc"].Get(i)
	return -(x*x*x - x)
}

// solve solves the Allan-Cahn equation
func solve(dt float64, solverName string) []complex128 {
	nx := 32
	ny := 32
	domainSize := []int{nx, ny}
	model := pf.NewModel()
	conc := pf.NewField("conc", nx*ny, nil)

	// Initialize with random concentration
	r := rand.New(rand.NewSource(0))
	for i := range conc.Data {
		conc.Data[i] = complex(2.0*r.Float64()-1.0, 0.0)
	}
	fmt.Printf("%f\n", maxReal(conc.Data))

	// Add constants
	gamma := pf.NewScalar("gamma", complex(0.02, 0.0)) // Gradient coefficient
	model.AddScalar(gamma)

	// Initialize the center
	model.AddField(conc)
	model.RegisterFunction("CHEMPOT", DerivChemPot)
	model.AddEquation("dconc/dt = CHEMPOT + gamma*LAP conc")

	// Initialize solver0.999506
	solver := pf.NewSolver(&model, domainSize, dt)

	if solverName == "implicitEuler" {
		solver.Stepper = &pf.ImplicitEuler{
			Dt: dt,
			FT: pf.NewFFTW(domainSize),
		}
	} else {
		solver.SetStepper(solverName)
	}
	solver.Solve(1, 100)
	return conc.Data
}

func hasNaN(data []complex128) bool {
	for _, v := range data {
		re := real(v)
		im := imag(v)
		if math.IsNaN(re) || math.IsNaN(im) {
			return true
		}
	}
	return false
}

func maxReal(data []complex128) float64 {
	maxval := 0.0
	for _, v := range data {
		if math.Abs(real(v)) > maxval {
			maxval = math.Abs(real(v))
		}
	}
	return maxval
}

func main() {
	timesteps := []float64{0.1, 1.0, 1.5, 1.9, 2.1}
	solvers := []string{"euler", "rk4", "implicitEuler"}

	for _, solver := range solvers {
		for _, dt := range timesteps {
			res := solve(dt, solver)
			stable := !hasNaN(res)
			fmt.Printf("Solver: %10s. Dt: %5.5f, Stable: %t\n", solver, dt, stable)
		}
	}
}
