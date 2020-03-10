package pf

// VolumeConservingLP is a term that can be added to a PDE such that the volume averaged
// integral of a field is zero. It does so by tuning a time dependent Lagrange multiplier.
//
// Note that due the way the multiplier is updated, the field is only approximately
// conserved. This is most easily seen when the fields changes rapidly
type VolumeConservingLP struct {
	Multiplier      float64
	Indicator       string
	Field           string
	CurrentIntegral float64
	Dt              float64
	NumNodes        int
	IsFirstUpdate   bool
}

// Construct build the function needed to evaluate the term
func (v *VolumeConservingLP) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			field[i] = bricks[v.Indicator].Get(i) * complex(v.Multiplier, 0.0)
		}
	}
}

// OnStepFinished updates the Lagrange multiplier
func (v *VolumeConservingLP) OnStepFinished(t float64, bricks map[string]Brick) {
	fieldIntegral := 0.0
	for i := 0; i < v.NumNodes; i++ {
		fieldIntegral += real(bricks[v.Field].Get(i))
	}

	// The indicator should be a derived field. They are not inverse FFT. Thus,
	// the volume integral of a derived field is the zero component of the fourier
	// transform
	indicatorIntegral := real(bricks[v.Indicator].Get(0))

	if v.IsFirstUpdate {
		v.CurrentIntegral = fieldIntegral
		v.IsFirstUpdate = false
	} else {
		deltaI := fieldIntegral - v.CurrentIntegral
		v.CurrentIntegral = fieldIntegral
		v.Multiplier = v.Multiplier - deltaI/(v.Dt*indicatorIntegral)
	}
}

// NewVolumeConservingLP returns a new instance of the volume conserving LP
func NewVolumeConservingLP(fieldName string, indicator string, dt float64, numNodes int) VolumeConservingLP {
	return VolumeConservingLP{
		Field:         fieldName,
		Indicator:     indicator,
		Dt:            dt,
		IsFirstUpdate: true,
		NumNodes:      numNodes,
	}
}
