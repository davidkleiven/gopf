package pf

// Euler performs semi-implicit euler method
type Euler struct {
	Dt     float64
	FT     FourierTransform
	Filter ModalFilter
}

// Step performs one euler step. If the equation is given by
// dy/dt = A*y + N(y), where N(y) is some non-linear function of y
// y_{n+1} = (y_n + dt*N(y_n))/(1 - dt*A)
func (eu *Euler) Step(m *Model) {
	cDt := complex(eu.Dt, 0.0)
	m.SyncDerivedFields()
	for _, f := range m.Fields {
		eu.FT.FFT(f.Data)
	}
	for _, f := range m.DerivedFields {
		eu.FT.FFT(f.Data)
	}

	for i := range m.Fields {
		rhs := m.GetRHS(i, eu.FT.Freq, 0.0)
		denum := m.GetDenum(i, eu.FT.Freq, 0.0)
		d := m.Fields[i].Data
		// Apply semi implicit scheme
		for j := range d {
			d[j] = (d[j] + cDt*rhs[j]) / (complex(1.0, 0.0) - cDt*denum[j])
		}

		if eu.Filter != nil {
			ApplyModalFilter(eu.Filter, eu.FT.Freq, d)
		}
	}

	// Inverse FFT
	for _, f := range m.Fields {
		eu.FT.IFFT(f.Data)
		DivRealScalar(f.Data, float64(len(f.Data)))
	}
}

// Propagate performs nsteps timesteps
func (eu *Euler) Propagate(nsteps int, m *Model) {
	for i := 0; i < nsteps; i++ {
		eu.Step(m)
	}
}

// SetFilter set a new modal filter
func (eu *Euler) SetFilter(filter ModalFilter) {
	eu.Filter = filter
}
