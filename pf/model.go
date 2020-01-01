package pf

// Field is a type that is used to represent a field in the context of phase field
// models
type Field struct {
	Data []complex128
	Name string
}

// Get returns the value at position i
func (f Field) Get(i int) complex128 {
	return f.Data[i]
}

// Brick is a generic interface to terms in a PDE
type Brick interface {
	Get(i int) complex128
}

// Scalar represents a scalar value
type Scalar struct {
	Value complex128
	Name  string
}

// SetFloat sets a new value
func (s Scalar) SetFloat(v float64) {
	s.Value = complex(v, 0.0)
}

// Get returns the scalar fvalue
func (s Scalar) Get(i int) complex128 {
	return s.Value
}

// Model is a type used to represent a general set of equations
type Model struct {
	Fields    []Field
	Bricks    map[string]Brick
	Equations []string
}

// AddField adds a field to the model
func (m Model) AddField(f Field) {
	m.Fields = append(m.Fields, f)
	m.Bricks[f.Name] = &f
}

// AddScalar adds a scalar to the model
func (m Model) AddScalar(s Scalar) {
	m.Bricks[s.Name] = &s
}

// AddEquation adds equations to the model
func (m Model) AddEquation(eq string) {
	m.Equations = append(m.Equations, eq)
}
