package main

import "github.com/davidkleiven/gopf/pf"

import "golang.org/x/exp/rand"

// Define some model constants
const c0 = 0.05   // Preferred concentration in the first phase
const c1 = 0.95   // Preferred concentration in the second phase
const l = 2.0     // Interface thickness
const w = 1.0     // Landau-barrier
const f0 = 1.0    // Curvature of the free energies
const m = 1.0     // Mobility
const dt = 0.01   // Timestep
const nsteps = 10 // Number of steps per epoch
const nepoch = 3  // Number of epochs

// H return the value of the interpolating function
func H(i int, bricks map[string]pf.Brick) complex128 {
	x := real(bricks["phi"].Get(i))
	val := 0.0
	if x > 1.0 {
		val = 1.0
	} else if x > 0.0 {
		val = 3.0*x*x - 2.0*x*x*x
	}
	return complex(val, 0.0)
}

// Hderiv return the derivative of H
func Hderiv(phi float64) float64 {
	val := 0.0
	if phi > 0.0 && phi < 1.0 {
		val = 6.0*phi - 6.0*phi*phi
	}
	return val
}

// Gderiv is the derivative of the Landau function phi^2*(1-phi)^2
func Gderiv(phi float64) float64 {
	val := 0.0
	if phi > 0.0 && phi < 1.0 {
		val = 2.0*phi - 6.0*phi*phi + 4.0*phi*phi*phi
	}
	return w * val
}

// PhiRHS calculates the right hand side of the phi equation
func PhiRHS(i int, bricks map[string]pf.Brick) complex128 {
	phi := real(bricks["phi"].Get(i))
	conc := real(bricks["conc"].Get(i))

	hprime := Hderiv(phi)
	dc := c0 - c1
	h := real(H(i, bricks))
	val := Gderiv(phi) + f0*hprime*dc*(conc-c0+h*dc)
	return complex(-val, 0.0)
}

func main() {
	N := 128
	conc := pf.NewField("conc", N*N, nil)
	phi := pf.NewField("phi", N*N, nil)

	// Initialize with a random concentration
	for i := range conc.Data {
		conc.Data[i] = complex(rand.Float64(), 0.0)
		phi.Data[i] = conc.Data[i]
	}

	model := pf.NewModel()
	model.AddField(conc)
	model.AddField(phi)
	concDiff := pf.NewScalar("c0minusc1", complex(c0-c1, 0.0))
	twof0M := pf.NewScalar("twof0M", complex(2.0*f0*m, 0.0))
	gamma := pf.NewScalar("gamma", complex(2.0*w/l, 0.0))
	model.AddScalar(concDiff)
	model.AddScalar(twof0M)
	model.AddScalar(gamma)

	model.RegisterFunction("INTERPOLANT", H)
	model.RegisterFunction("PHI_RHS", PhiRHS)

	model.AddEquation("dconc/dt = twof0M*LAP conc + LAP*twof0M*c0minusc1*INTERPOLANT")
	model.AddEquation("dphi/dt = PHI_RHS + gamma*LAP*phi")

	solver := pf.NewSolver(&model, []int{N, N}, dt)

	fileSaver := pf.NewFloat64IO("kks")
	solver.AddCallback(fileSaver.SaveFields)

	model.Summarize()
	solver.Solve(nepoch, nsteps)
	pf.WriteXDMF("kks.xdmf", []string{"conc", "phi"}, "kks", nepoch, []int{N, N})
}
