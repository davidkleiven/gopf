package pf

import (
	"encoding/json"
	"fmt"

	"github.com/davidkleiven/gosfft/sfft"
)

// SolverCB is function type that can be added to the solver it is executed after each
// iteration
type SolverCB func(s *Solver, epoch int)

// FourierTransform is a type used to represent fourier transforms
type FourierTransform interface {
	FFT(data []complex128) []complex128
	IFFT(data []complex128) []complex128
	Freq(i int) []float64
}

// Solver is a type used to solve phase field equations
type Solver struct {
	Model     *Model
	FT        FourierTransform
	Dt        float64
	Callbacks []SolverCB
	Monitors  []PointMonitor
}

// NewSolver initializes a new solver
func NewSolver(m *Model, domainSize []int, dt float64) *Solver {
	var solver Solver
	m.Init()
	solver.Model = m
	solver.Dt = dt
	solver.Callbacks = []SolverCB{}
	solver.Monitors = []PointMonitor{}

	if len(domainSize) == 2 {
		solver.FT = sfft.NewFFT2(domainSize[0], domainSize[1])
	} else if len(domainSize) == 3 {
		solver.FT = sfft.NewFFT3(domainSize[0], domainSize[1], domainSize[2])
	} else {
		panic("solver: Domain size has to be an array of length 2 (2D calculation) or 3 (3D calculation)")
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
	cDt := complex(s.Dt, 0.0)
	for i := 0; i < nsteps; i++ {
		s.Model.SyncDerivedFields()
		for _, f := range s.Model.Fields {
			s.FT.FFT(f.Data)
		}
		for _, f := range s.Model.DerivedFields {
			s.FT.FFT(f.Data)
		}

		for i := range s.Model.Fields {
			rhs := s.Model.GetRHS(i, s.FT.Freq, 0.0)
			denum := s.Model.GetDenum(i, s.FT.Freq, 0.0)
			d := s.Model.Fields[i].Data
			// Apply semi implicit scheme
			for j := range d {
				d[j] = (d[j] + cDt*rhs[j]) / (complex(1.0, 0.0) - cDt*denum[j])
			}
		}

		// Inverse FFT
		for _, f := range s.Model.Fields {
			s.FT.IFFT(f.Data)
			DivRealScalar(f.Data, float64(len(f.Data)))
		}

		for i := range s.Model.UserDef {
			s.Model.UserDef[i].OnStepFinished(0.0, s.Model.Bricks)
		}
	}
}

// Solve solves the equation
func (s *Solver) Solve(nepochs int, nsteps int) {
	for i := 0; i < nepochs; i++ {
		fmt.Printf("Epoch %5d of %5d\n", i, nepochs)
		s.Propagate(nsteps)

		for _, cb := range s.Callbacks {
			cb(s, i)
		}

		// Update monitors
		for i, m := range s.Monitors {
			s.Monitors[i].Add(real(s.Model.Bricks[m.Field].Get(m.Site)))
		}
	}
}

// AddMonitor adds a new monitor to the solver
func (s *Solver) AddMonitor(m PointMonitor) {
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
