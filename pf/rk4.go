package pf

import "github.com/davidkleiven/gopf/pfutil"

// RK4 implements the fourth order Runge-Kutta scheme. Dt is the timestep
// FT is a fourier transform object used to translate back-and fourth between
// fourier domain.
type RK4 struct {
	Dt          float64
	FT          FourierTransform
	Filter      ModalFilter
	CurrentStep int
}

// Step performs one RK4 time step. If the equation is given by
// dy/dt = A*y + N(y)
// where N(y) is some non-linear function, the update consists
// of the following steps
// y_{n+1} = (y_n + dt*(k1 + 2*k2 + 2*k3 + k4)/6)/(1 - dt*A),
// where the k coefficients are given by
//
// k1 = N(y_n)
// k2 = N( (y_n + dt*k1/2)/(1 - 0.5*dt*A) )
// k3 = N( (y_n + dt*k2/2)/(1 - 0.5*dt*A) )
// k4 = N( (y_n + dt*k3)/(1 - dt*A) )
//
// This leads to a scheme that is first order accurate in dt, but with
// much better stability properties than the Euler scheme
func (rk *RK4) Step(m *Model) {
	m.SyncDerivedFields()
	cDt := complex(rk.Dt, 0.0)
	for _, f := range m.Fields {
		rk.FT.FFT(f.Data)
	}

	initial := make([]Field, len(m.Fields))
	final := make([]Field, len(m.Fields))
	kFactor := make([]Field, len(m.Fields))
	for i := range m.Fields {
		initial[i] = m.Fields[i].Copy()
		final[i] = m.Fields[i].Copy()
		kFactor[i] = m.Fields[i].Copy()
	}

	rk.firstCorrection(m, kFactor)
	rk.PrepareNextCorrection(initial, final, kFactor, m, 1.0/6.0)
	rk.correction(m, kFactor, 0.5)
	rk.PrepareNextCorrection(initial, final, kFactor, m, 1.0/3.0)
	rk.correction(m, kFactor, 0.5)
	rk.PrepareNextCorrection(initial, final, kFactor, m, 1.0/3.0)
	rk.correction(m, kFactor, 1.0)
	rk.PrepareNextCorrection(initial, final, kFactor, m, 1.0/6.0)

	// NOTE: If there are implicit terms, the scheme is only accurate to
	// first order. But the stability of RK4 should be better than the Euler
	// scheme
	t := rk.GetTime()
	for i := range m.Fields {
		denum := m.GetDenum(i, rk.FT.Freq, t)
		for j := range final[i].Data {
			final[i].Data[j] /= (complex(1.0, 0.0) - cDt*denum[j])
		}
		copy(m.Fields[i].Data, final[i].Data)

		if rk.Filter != nil {
			ApplyModalFilter(rk.Filter, rk.FT.Freq, m.Fields[i].Data)
		}
	}

	for _, f := range m.Fields {
		rk.FT.IFFT(f.Data)
		pfutil.DivRealScalar(f.Data, float64(len(f.Data)))
	}
}

// PrepareNextCorrection updates the final fields and rests the fields of the model to the original
func (rk *RK4) PrepareNextCorrection(initial []Field, final []Field, kFactor []Field, m *Model, factor float64) {
	for i := range final {
		for j := range final[i].Data {
			final[i].Data[j] += complex(factor*rk.Dt, 0.0) * kFactor[i].Data[j]
		}
		copy(m.Fields[i].Data, initial[i].Data)
	}
}

// Calculates the first correction factor
func (rk *RK4) firstCorrection(m *Model, kFactor []Field) {
	for _, f := range m.DerivedFields {
		rk.FT.FFT(f.Data)
	}

	t := rk.GetTime()
	for i := range m.Fields {
		kFactor[i].Data = m.GetRHS(i, rk.FT.Freq, t)
	}
}

// correction calculates a general RK correction. If the equation is given by
// dy/dt = A*y + N(t, y), this function returns N(t + factor*dt, (y + factor*kFactor)/(1 - factor*dt*A)).
// In RK4 factor=0.5 for the middle steps and 1 for the last
func (rk *RK4) correction(m *Model, kFactor []Field, factor float64) {
	t := rk.GetTime()
	for i, f := range m.Fields {
		denum := m.GetDenum(i, rk.FT.Freq, t)
		for j := range f.Data {
			f.Data[j] += complex(factor*rk.Dt, 0.0) * kFactor[i].Data[j]

			// NOTE: We use backward scheme for the linear part. This leads to a
			// first order accurate scheme
			f.Data[j] /= (complex(1.0, 0.0) - complex(factor*rk.Dt, 0.0)*denum[j])
		}
		rk.FT.IFFT(f.Data)
		pfutil.DivRealScalar(f.Data, float64(len(f.Data)))
	}
	m.SyncDerivedFields()

	for _, f := range m.Fields {
		rk.FT.FFT(f.Data)
	}
	for _, f := range m.DerivedFields {
		rk.FT.FFT(f.Data)
	}

	for i := range m.Fields {
		kFactor[i].Data = m.GetRHS(i, rk.FT.Freq, t)
	}
}

// Propagate evolves the fields a given number of steps
func (rk *RK4) Propagate(nsteps int, m *Model) {
	for i := 0; i < nsteps; i++ {
		rk.Step(m)
		rk.CurrentStep++
	}
}

// SetFilter sets a new modal filter
func (rk *RK4) SetFilter(filter ModalFilter) {
	rk.Filter = filter
}

// GetTime returns the current time
func (rk *RK4) GetTime() float64 {
	return float64(rk.CurrentStep) * rk.Dt
}
