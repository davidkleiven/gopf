package pf

// UserDefinedTerm is an interface that can be used to add
// special terms to a PDE that cannot easily be defined via a string
// representation
type UserDefinedTerm interface {
	// Construct creates the righ hand side function for the term
	Construct(bricks map[string]Brick) Term

	// OnStepFinished is gets called after each time step. It can be used
	// to do nessecary updates after the fields have been updated. If no such
	// updates are nessecary, just implement an empty function
	OnStepFinished(t float64, bricks map[string]Brick)
}

// GenericFunction is a generic function that may depend
// on any of the fields
type GenericFunction func(i int, bricks map[string]Brick) complex128

// GenericFunctionTerm is a small struct used to represent user defined functions
type GenericFunctionTerm struct {
	Name string
}

// Construct creates the proper RHS function
func (g *GenericFunctionTerm) Construct(bricks map[string]Brick) Term {
	lap := LaplacianN{Power: 1}
	return func(freq Frequency, t float64, field []complex128) []complex128 {
		for i := range field {
			field[i] = bricks[g.Name].Get(i)
		}

		if g.Name[:3] == "LAP" {
			lap.Eval(freq, field)
		}
		return field
	}
}

// OnStepFinished does nothing
func (g *GenericFunctionTerm) OnStepFinished(t float64, bricks map[string]Brick) {}
