package main

import (
	"math"

	"github.com/davidkleiven/gopf/elasticity"
	"github.com/davidkleiven/gopf/pf"
	"gonum.org/v1/gonum/mat"
)

// H is the function used to interpolate between the phases
func H(x float64) float64 {
	return 3.0*x*x - 2.0*x*x*x
}

// dHdx is the the derivative of x
func dHdx(x float64) float64 {
	return 6.0*x - 6.0*x*x
}

// Landau is the function used to create two distinct minima
func Landau(x float64) float64 {
	return x*x - 2.0*x*x*x + x*x*x*x
}

// dLandaudx is the derivative of the landau function
func dLandaudx(x float64) float64 {
	return 2.0*x - 6.0*x*x + 4.0*x*x*x
}

// ModelFunctions is a struct that is used to define the nessecary functions for this
// model
type ModelFunctions struct {
	A float64
	B float64
	W float64
	M float64
}

// ChemicalPotential returns the derivative of the free energy with respect
// to the concentration
func (m *ModelFunctions) ChemicalPotential(i int, bricks map[string]pf.Brick) complex128 {
	conc := real(bricks["conc"].Get(i))
	x := real(bricks["phase"].Get(i))
	dfdc := complex(m.A*conc*(1.0-H(x))-m.B*(1.0-conc)*H(x), 0.0) * complex(m.M, 0.0)
	return -dfdc
}

// DerivPhase returns the derivative with respect to the phase order parameter
func (m *ModelFunctions) DerivPhase(i int, bricks map[string]pf.Brick) complex128 {
	conc := real(bricks["conc"].Get(i))
	x := real(bricks["phase"].Get(i))
	val := -0.5*m.A*math.Pow(conc, 2)*dHdx(x) + 0.5*m.B*math.Pow(1.0-conc, 2)*dHdx(x) + m.W*dLandaudx(x)
	return -complex(val, 0.0)
}

// SmearingDeriv is used to keep the volume of the precipitate constant
func SmearingDeriv(i int, bricks map[string]pf.Brick) complex128 {
	x := real(bricks["phase"].Get(i))
	return complex(dHdx(x), 0.0)
}

func main() {
	dt := 0.1
	M := 64
	domainSize := []int{M, M}
	N := pf.ProdInt(domainSize)
	model := pf.NewModel()
	conc := pf.NewField("conc", N, nil)
	phase := pf.NewField("phase", N, nil)

	// Initialize fields (square at the center)
	for i := 0; i < N; i++ {
		r := i / M
		c := i % M
		if r > 3*M/8 && r < 5*M/8 && c > 3*M/8 && c < 5*M/8 {
			conc.Data[i] = complex(1.0, 0.0)
			phase.Data[i] = complex(1.0, 0.0)
		}
	}
	kappa := pf.NewScalar("kappa", complex(0.1, 0.0))
	model.AddScalar(kappa)

	model.AddField(conc)
	model.AddField(phase)

	mf := ModelFunctions{
		A: 0.1,
		B: 0.1,
		W: 0.1,
		M: 1.0,
	}

	// Register the two functions that is required to evolve the system
	model.RegisterFunction("CHEMICALPOT", mf.ChemicalPotential)
	model.RegisterFunction("DERIV_PHASE_ORDER", mf.DerivPhase)
	model.RegisterFunction("SMEARING_DERIV", SmearingDeriv)

	// Add volume conserving constraint
	volPhase := pf.NewVolumeConservingLP("phase", "SMEARING_DERIV", dt, N)
	model.RegisterUserDefinedTerm("CONSERVE_PREC_VOL", &volPhase, nil)

	// Register the linear elasticity term
	misfit := mat.NewDense(3, 3, []float64{0.05, 0.0, 0.0, 0.0, -0.01, 0.0, 0.0, 0.0, 0.0})
	matProp := elasticity.CubicMaterial(110.0, 60.0, 30.0)
	linelast := pf.NewHomogeneousModolus("phase", domainSize, matProp, misfit)
	model.RegisterUserDefinedTerm("LIN_ELAST", linelast, nil)

	model.AddEquation("dconc/dt = CHEMICALPOT + kappa*LAP conc")
	model.AddEquation("dphase/dt = DERIV_PHASE_ORDER + LIN_ELAST + kappa*LAP phase + CONSERVE_PREC_VOL")

	// Initialize the solver
	solver := pf.NewSolver(&model, domainSize, dt)

	// Initialize uint8 IO
	out := pf.NewFloat64IO("precipitate2D")
	solver.AddCallback(out.SaveFields)

	solver.Solve(10, 10)

	// Write XDMF
	pf.WriteXDMF("precipitate2D.xdmf", []string{"conc", "phase"}, "precipitate2D", 10, domainSize)
}
