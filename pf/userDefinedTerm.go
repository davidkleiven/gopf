package pf

// PureTerm is an interface that can be used to add
// special terms to a PDE that cannot easily be defined via a string
// representation. This term type is either linear or non-linear.
// Since, a linear term can be treated as a non-linear term, it is
// beneficial to treat the linear and non-linear part differently.
// Terms that exhibit both linear and non-linear parts should be
// treated via the MixedTerm interface. In this context a linear term
// is a term on the form f({otherfields})*field, where f is a function
// (or operator) that is independent of the field in question, but can
// depend on the other fields in the model. Terms of this form can easily be treated
// implicitly.
//
// Non-linear terms, are terms on the form f({all fields}), where f is a function
// (or operator) that depends on all the fields. Terms on this form will be
// treated explicitly when evolving the fields
type PureTerm interface {
	// Construct creates the righ hand side function for the term.
	// The returned function should should give the fourier transformed
	// quantity. If this method is a linear term, it should return the
	// fourier transform of f({otherfields}) (excluding the multiplication
	// with the field in question). If it is a non-linear term, it should
	// return the fourier transform of f({allfields}). See the documentation
	// of the interface for a detailed definition of f({otherfields}) and
	// f({allfields}). When bricks is passed to this method, all fields
	// have already been fourier transformed.
	Construct(bricks map[string]Brick) Term

	// OnStepFinished is gets called after each time step. It can be used
	// to do nessecary updates after the fields have been updated. If no such
	// updates are nessecary, just implement an empty function
	OnStepFinished(t float64, bricks map[string]Brick)
}

// MixedTerm is a type that can be used to represents terms that have both
// a linear part and a non-linear part. Mixed terms are on the form
// f({otherfields})*field + g({allfields}), where f is a function (or operator)
// that depends on all the other fields in the model and g is a function/operator
// that depends on all the fields in the model
type MixedTerm interface {
	// ConstructLinear builds the function to evaluate the linear part of the
	// term. The function returned should give the fourier transform of
	// f({otherfields}). The bricks parameter contains the fourier transform
	// of all bricks
	ConstructLinear(bricks map[string]Brick) Term

	// ConstructNonLinear returns a function that calculates the non-linear
	// part of the expression. The function being returned should calculate
	// the fourier transform of f({allfields}). The bricks parameter contains
	// the fourier transform of all known bricks
	ConstructNonLinear(bricks map[string]Brick) Term

	// OnStepFinished is called after each step is finished. If a term needs
	// be updated based on how the fields evolves, the update should happen
	// inside this method
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
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			field[i] = bricks[g.Name].Get(i)
		}

		if g.Name[:3] == "LAP" {
			lap.Eval(freq, field)
		}
	}
}

// OnStepFinished does nothing
func (g *GenericFunctionTerm) OnStepFinished(t float64, bricks map[string]Brick) {}
