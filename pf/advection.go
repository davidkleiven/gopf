package pf

import "fmt"

// Advection is a type that is used to build. This represents a term of the form
// v dot grad field, where v is a velocity vector and field is the name of a field
// Since terms in GOPF are assumed to enter on the right hand side of equations of
// the form dy/dt = ..., this term return -v dot grad field.
// Field is the name of the field and VelocifyFields is a slice with the names
// of the velocity components. The number of velocity components has to be exactly
// equal to the number of dimensions (e.g. if it is 2D model then the length of this
// slice should be two)
type Advection struct {
	Field          string
	VelocityFields []string
}

// GradName returns the gradient term associated with the field
func (ad *Advection) GradName(comp int) string {
	return fmt.Sprintf("%s_%d", ad.Field, comp)
}

// GetName returns the name of the brick corresponding to
// vDotGradField
func (ad *Advection) GetName() string {
	name := ""
	for _, vel := range ad.VelocityFields {
		name += vel
	}
	name += "Dot" + "Grad" + ad.Field
	return name
}

// AllFieldsExist returns true if all the fields are registered in the model
func (ad *Advection) AllFieldsExist(m *Model) bool {
	for _, vf := range ad.VelocityFields {
		if !m.IsBrickName(vf) {
			return false
		}
	}
	return m.IsBrickName(ad.Field)
}

// PrepareModel adds the nessecary fields to the model in order to be
// able to use the Advection term
func (ad *Advection) PrepareModel(N int, m *Model, FT FourierTransform) {
	dim := len(FT.Freq(0))
	if len(ad.VelocityFields) != dim {
		panic("Advection: Inconsistent number of velocity fields")
	}

	if !ad.AllFieldsExist(m) {
		panic("Advection: Make sure that field and all the velocity fields are added to the model")
	}

	for d := 0; d < dim; d++ {
		grad := GradientCalculator{
			FT:   FT,
			Comp: d,
		}
		dField := grad.ToDerivedField(ad.GradName(d), N, m.Bricks[ad.Field])
		m.RegisterDerivedField(dField)
	}

	vDotGradField := DerivedField{
		Name: ad.GetName(),
		Data: make([]complex128, N),
		Calc: func(data []complex128) {
			dim := len(ad.VelocityFields)
			Clear(data)
			for d := 0; d < dim; d++ {
				vBrick := m.Bricks[ad.VelocityFields[d]]
				gBrick := m.Bricks[ad.GradName(d)]
				for i := range data {
					data[i] += vBrick.Get(i) * gBrick.Get(i)
				}
			}
		},
	}
	m.RegisterDerivedField(vDotGradField)
}

// Construct builds the right hand side function
func (ad *Advection) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		brick := bricks[ad.GetName()]
		for i := range field {
			field[i] = -brick.Get(i)
		}
	}
}

// OnStepFinished does nothing as we don't need any updates in between steps
func (ad *Advection) OnStepFinished(t float64) {}
