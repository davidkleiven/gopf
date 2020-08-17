package pf

import (
	"log"
	"math/cmplx"

	"github.com/davidkleiven/gononlin/nonlin"
	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/exp/linsolve"
)

// ImplicitEuler performs a full implicit euler method. We have a problem
// of the form dy/dt = Ly + N(y), where L is a linear operator and N is a non-linear
// operator. The implicit update scheme is
// y_{n+1} = y_n + dt*Ly_{n+1} + N(y_{n+1}), this gives rise to a non-linear set
// of equations that must be solved on each time step. Provided that the non-linear
// solver is able to find the solution, this scheme is stable for all time steps dt.
// However, in practice the non-linear solver converges faster it dt is small. Thus,
// if the solution changes rapidly it might be wise to use a small time step.
type ImplicitEuler struct {
	Dt          float64
	FT          *pfutil.FFTWWrapper
	Filter      ModalFilter
	CurrentStep int

	// each time step. If not given (or nil), a solver with sensible default values
	// will be used
	NonlinSolver *nonlin.NewtonKrylov
}

func (ie *ImplicitEuler) isImag(f []float64) bool {
	tol := 1e-10
	for i := range f {
		if f[i] > tol && f[i] < 0.5-tol {
			return true
		}
	}
	return false
}

// fields2vec transfer all the fourier transformed fields into out.
// The real and imaginary part of a fourier amplitude is stored as two
// successive items in out. However, since it is assumed that that the field
// is real, there are some symmetries in the fourier transform (e.g. A(-k) = A(k)^*)
// Thus, if there are N fields and M nodes, the length of out should be N*M
func (ie *ImplicitEuler) fields2vec(fields []Field, out []float64) {
	for i, f := range fields {
		ie.fieldData2vec(f.Data, out[i*len(f.Data):])
	}
}

func (ie *ImplicitEuler) fieldData2vec(data []complex128, out []float64) {
	for i := range data {
		out[i] = real(data[i])
	}
}

func (ie *ImplicitEuler) vec2fields(vec []float64, fields []Field) {
	counter := 0
	for i := range fields {
		for j := 0; j < len(fields[i].Data); j++ {
			fields[i].Data[j] = complex(vec[counter], 0.0)
			counter++
		}
	}
}

func (ie *ImplicitEuler) updateEquation(newFields []float64, out []float64, rhsPrev []complex128, origFields []Field, m *Model) {
	// Transfer the fields in the work array model
	ie.vec2fields(newFields, m.Fields)
	//ie.ifft(m) // Inverse fourier transform the new fields
	ie.fft(m) // Update all derived fields and fourier transform everything

	t := ie.GetTime()
	cDt := complex(ie.Dt, 0.0)
	for i := range m.Fields {
		N := len(m.Fields[i].Data)
		rhs := m.GetRHS(i, ie.FT.Freq, t)
		denum := m.GetDenum(i, ie.FT.Freq, t)

		// Overwrite rhs with the complex the equation data
		for j := range rhs {
			//update := (origFields[i].Data[j] + cDt*rhs[j]) / (1.0 - cDt*denum[j])
			factor := cmplx.Exp(denum[j] * cDt)
			integral := ie.nonlinearIntegral(denum[j], rhs[j], rhsPrev[i*N+j])
			//update := origFields[i].Data[j]*factor + 0.5*cDt*(rhsPrev[i*N+j]*factor+rhs[j])
			update := origFields[i].Data[j]*factor + integral
			rhs[j] = m.Fields[i].Data[j] - update
		}
		ie.FT.IFFT(rhs)
		pfutil.DivRealScalar(rhs, float64(N))
		ie.fieldData2vec(rhs, out[i*N:])
	}
}

// GetTime returns the current time
func (ie *ImplicitEuler) GetTime() float64 {
	return ie.Dt * float64(ie.CurrentStep)
}

// fft performs forward FFT on all fields in the model
func (ie *ImplicitEuler) fft(m *Model) {
	m.SyncDerivedFields()
	for _, f := range m.Fields {
		ie.FT.FFT(f.Data)
	}
	for _, f := range m.DerivedFields {
		ie.FT.FFT(f.Data)
	}
}

// ifft performs inverse fourier transform on all fields
func (ie *ImplicitEuler) ifft(m *Model) {
	for _, f := range m.Fields {
		ie.FT.IFFT(f.Data)
		pfutil.DivRealScalar(f.Data, float64(len(f.Data)))
	}
}

func (ie *ImplicitEuler) homotopyUpdate(originalFields []Field, rhsPrev []complex128, m *Model, lamb float64) {
	t := ie.GetTime()
	cDt := complex(ie.Dt, 0.0)
	ie.ifft(m)

	// Updates the derived fields and fourier transforms
	// all the fields + the derived fields
	ie.fft(m)

	nlWeight := complex(lamb, 0.0)
	for i := range m.Fields {
		rhs := m.GetRHS(i, ie.FT.Freq, t)
		denum := m.GetDenum(i, ie.FT.Freq, t)

		for j := range denum {
			factor := cmplx.Exp(denum[j] * cDt)
			integral := ie.nonlinearIntegral(denum[j], rhs[j], rhsPrev[i*len(rhs)+j])
			//m.Fields[i].Data[j] = originalFields[i].Data[j]*factor + 0.5*nlWeight*cDt*(rhs[j]+factor*rhsPrev[i*len(rhs)+j])
			m.Fields[i].Data[j] = originalFields[i].Data[j]*factor + nlWeight*integral
		}
	}
}

func (ie *ImplicitEuler) defaultNumHomotopy() int {
	return 10
}

func (ie *ImplicitEuler) defaultTol() float64 {
	return 1e-10
}

func (ie *ImplicitEuler) nonlinearIntegral(denum complex128, rhs complex128, rhsPrev complex128) complex128 {
	cDt := complex(ie.Dt, 0.0)
	tol := 1e-5
	a := rhsPrev
	b := (rhs - rhsPrev) / cDt
	f := cmplx.Exp(denum * cDt)

	if cmplx.Abs(denum) < tol {
		return 0.5 * cDt * (rhs + rhsPrev*f)
	}
	return a*(f-1.0)/denum + b*(f-denum*cDt-1.0)/(denum*denum)
}

// Step evolves the equation one timestep
func (ie *ImplicitEuler) Step(m *Model) {
	numNodes := len(m.Fields[0].Data)
	t := ie.GetTime()
	x0 := make([]float64, len(m.Fields)*numNodes)

	origFields := make([]Field, len(m.Fields)) // Fourier transformed initial fields

	rhsPrev := make([]complex128, len(x0)) // Right hand side of the equations at the initial
	ie.fft(m)
	for i := range m.Fields {
		origFields[i] = m.Fields[i].Copy()
		copy(rhsPrev[i*numNodes:], m.GetRHS(i, ie.FT.Freq, t)) // Fourier transformed rhs
	}
	ie.ifft(m)

	// Perform one explicit euler step and use that as the initial guess
	explicitEuler := Euler{
		Dt:          ie.Dt,
		FT:          ie.FT,
		Filter:      ie.Filter,
		CurrentStep: ie.CurrentStep,
	}
	explicitEuler.Step(m)
	ie.fields2vec(m.Fields, x0)

	problem := nonlin.Problem{
		F: func(out []float64, x []float64) {
			ie.updateEquation(x, out, rhsPrev, origFields, m)
		},
	}

	if ie.NonlinSolver == nil {
		nonlinSolver := DefaultNonLinSolver()
		ie.NonlinSolver = &nonlinSolver
	}

	res := ie.NonlinSolver.Solve(problem, x0)
	if !res.Converged {
		log.Printf("Warning: Iterative solver did not converge\n")
	}
	ie.vec2fields(res.X, m.Fields)
	ie.CurrentStep++
}

// SetFilter sets a new filter. Currently this has no effect in this
// timestepper
func (ie *ImplicitEuler) SetFilter(f ModalFilter) {
	ie.Filter = f
}

// DefaultNonLinSolver returns the default non-linear solver used in Implicit Euler.
// Internally, the GMRES method is used. With the default settings, this method can
// consume a lot of memory depending on the problem. If the program uses too much memory,
// try to use a restarted version of GMRES (e.g. InnerMethod: &linsolve.GMRES{Restart: 50})
// See GMRES description at https://godoc.org/github.com/gonum/exp/linsolve for further
// details
func DefaultNonLinSolver() nonlin.NewtonKrylov {
	return nonlin.NewtonKrylov{
		Maxiter:     50,
		StepSize:    1e-3,
		Tol:         1e-7,
		Stencil:     6,
		InnerMethod: &linsolve.GMRES{},
	}
}
