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
	OnStepFinished(t float64)
}
