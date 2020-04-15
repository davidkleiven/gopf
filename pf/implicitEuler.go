package pf

import (
	"fmt"
	"math/cmplx"

	"github.com/davidkleiven/gononlin/nonlin"
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
	FT          *FFTWWrapper
	Filter      ModalFilter
	CurrentStep int

	// NonlinSolver used to solve the non-linear set of equation that emerges on
	// each time step. If not given (or nil), a solver with sensible default values
	// will be used
	NonlinSolver *nonlin.NewtonBCGS
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
	iterator := UniqueFreqIterator{
		Freq: ie.FT.Freq,
		End:  len(data),
	}
	counter := 0
	for j := iterator.Next(); j != -1; j = iterator.Next() {
		freq := ie.FT.Freq(j)
		if ie.isImag(freq) {
			out[counter] = real(data[j])
			out[counter+1] = imag(data[j])
			counter += 2
		} else {
			out[counter] = real(data[j])
			counter++
		}
	}
}

// vec2fields is the inverse operation of fields2vec. The value in vec is transferred
// into fields
func (ie *ImplicitEuler) vec2fields(vec []float64, fields []Field) {
	counter := 0
	for i := range fields {
		iterator := UniqueFreqIterator{
			Freq: ie.FT.Freq,
			End:  len(fields[i].Data),
		}
		for j := iterator.Next(); j != -1; j = iterator.Next() {
			freq := ie.FT.Freq(j)
			if ie.isImag(freq) {
				fields[i].Data[j] = complex(vec[counter], vec[counter+1])
				counter += 2
			} else {
				fields[i].Data[j] = complex(vec[counter], 0.0)
				counter++
			}

			conj := ie.FT.ConjugateNode(j)
			fields[i].Data[conj] = cmplx.Conj(fields[i].Data[j])
		}
	}
}

func (ie *ImplicitEuler) updateEquation(newFields []float64, out []float64, rhsPrev []complex128, origFields []Field, m *Model) {
	// Transfer the fields in the work array model
	ie.vec2fields(newFields, m.Fields)
	ie.ifft(m) // Inverse fourier transform the new fields
	ie.fft(m)  // Update all derived fields and fourier transform everything

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
		DivRealScalar(f.Data, float64(len(f.Data)))
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
	x0 := make([]float64, len(m.Fields)*len(m.Fields[0].Data))
	rhsPrev := make([]complex128, len(x0))
	ie.fft(m)
	origFields := make([]Field, len(m.Fields))
	for i := range m.Fields {
		origFields[i] = m.Fields[i].Copy()
		copy(rhsPrev[i*numNodes:], m.GetRHS(i, ie.FT.Freq, t))
	}

	problem := nonlin.Problem{
		F: func(out []float64, x []float64) {
			ie.updateEquation(x, out, rhsPrev, origFields, m)
		},
	}

	if ie.NonlinSolver == nil {
		ie.NonlinSolver = &nonlin.NewtonBCGS{
			Maxiter:  500,
			StepSize: 1e-3 / ie.Dt,
			Tol:      1e-7,
			Stencil:  6,
		}

	}

	// Semi-implicit step serves as our initial guess
	ie.homotopyUpdate(origFields, rhsPrev, m, 1.0)
	ie.fields2vec(m.Fields, x0)

	res := ie.NonlinSolver.Solve(problem, x0)
	if !res.Converged {
		fmt.Printf("Warning: Iterative solver did not converge\n")
	}
	ie.vec2fields(res.X, m.Fields)
	ie.ifft(m)
	ie.CurrentStep++
}

// SetFilter sets a new filter. Currently this has no effect in this
// timestepper
func (ie *ImplicitEuler) SetFilter(f ModalFilter) {
	ie.Filter = f
}
