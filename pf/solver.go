package pf

import (
	"encoding/json"
	"fmt"
	"time"
)

// SolverCB is function type that can be added to the solver it is executed after each
// iteration
type SolverCB func(s *Solver, epoch int)

// TimeStepper is a generic interface for a the time stepper types
type TimeStepper interface {
	Step(m *Model)
	SetFilter(filter ModalFilter)
	GetTime() float64
}

// FourierTransform is a type used to represent fourier transforms
type FourierTransform interface {
	FFT(data []complex128) []complex128
	IFFT(data []complex128) []complex128
	Freq(i int) []float64
}

// Solver is a type used to solve phase field equations
type Solver struct {
	Model      *Model
	Dt         float64
	FT         FourierTransform
	Stepper    TimeStepper
	Callbacks  []SolverCB
	Monitors   []Monitor
	StartEpoch int
}

// NewSolver initializes a new solver
func NewSolver(m *Model, domainSize []int, dt float64) *Solver {
	var solver Solver
	m.Init()
	solver.Model = m
	solver.Dt = dt
	solver.Callbacks = []SolverCB{}
	solver.Monitors = []Monitor{}
	solver.FT = NewFFTW(domainSize)

	solver.Stepper = &Euler{
		Dt: solver.Dt,
		FT: solver.FT,
	}

	// Sanity check for fields
	N := ProdInt(domainSize)
	for _, f := range m.Fields {
		if len(f.Data) != N {
			panic("solver: Inconsistent domain size and number of grid points")
		}
	}
	return &solver
}

// AddCallback appends a new callback function to the solver
func (s *Solver) AddCallback(cb SolverCB) {
	s.Callbacks = append(s.Callbacks, cb)
}

// Propagate evolves the equation a fixed number of steps
func (s *Solver) Propagate(nsteps int) {
	for i := 0; i < nsteps; i++ {
		s.Stepper.Step(s.Model)
		t := s.Stepper.GetTime()
		for j := range s.Model.ImplicitTerms {
			s.Model.ImplicitTerms[j].OnStepFinished(t, s.Model.Bricks)
		}
		for j := range s.Model.ExplicitTerms {
			s.Model.ExplicitTerms[j].OnStepFinished(t, s.Model.Bricks)
		}
		for j := range s.Model.MixedTerms {
			s.Model.MixedTerms[j].OnStepFinished(t, s.Model.Bricks)
		}
	}
}

// SetStepper updates the stepper method based on a string.
// name has to be one of ["euler", "rk4"]
func (s *Solver) SetStepper(name string) {
	switch name {
	case "euler":
		s.Stepper = &Euler{
			Dt: s.Dt,
			FT: s.FT,
		}
	case "rk4":
		s.Stepper = &RK4{
			Dt: s.Dt,
			FT: s.FT,
		}
	default:
		panic("Unknown stepper scheme")
	}
}

// Solve solves the equation
func (s *Solver) Solve(nepochs int, nsteps int) {
	for i := 0; i < nepochs; i++ {
		now := time.Now()
		fmt.Printf("%02d:%02d:%02d - Epoch %5d of %5d\n", now.Hour(), now.Minute(), now.Second(), i+s.StartEpoch, nepochs+s.StartEpoch)
		s.Propagate(nsteps)

		for _, cb := range s.Callbacks {
			cb(s, i+s.StartEpoch)
		}

		// Update monitors
		for i := range s.Monitors {
			s.Monitors[i].Add(s.Model.Bricks)
		}
	}
}

// AddMonitor adds a new monitor to the solver
func (s *Solver) AddMonitor(m Monitor) {
	s.Monitors = append(s.Monitors, m)
}

// JSONifyMonitors return a JSON representation of all the monitors
func (s *Solver) JSONifyMonitors() []byte {
	res, err := json.Marshal(s.Monitors)
	if err != nil {
		panic(err)
	}
	return res
}
