package pf

import (
	"strings"
)

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

// NewField initializes a new field
func NewField(name string, N int, data []complex128) Field {
	var field Field
	if data == nil {
		field.Data = make([]complex128, N)
	} else {
		if len(data) != N {
			panic("model: Inconsistent length of data")
		}
		field.Data = data
	}
	field.Name = name
	return field
}

// DerivedFieldCalc is a function that calculates the derived field
type DerivedFieldCalc func(data []complex128)

// DerivedField is a type that is derived from Fields via multiplications
// and power operations
type DerivedField struct {
	Data []complex128
	Name string
	Calc DerivedFieldCalc
}

// Get returns the value at position i
func (d DerivedField) Get(i int) complex128 {
	return d.Data[i]
}

// Update recalculates the derived fields and places the result in Data
func (d *DerivedField) Update() {
	d.Calc(d.Data)
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

// NewScalar returns a new scalar value
func NewScalar(name string, value complex128) Scalar {
	return Scalar{
		Name:  name,
		Value: value,
	}
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
	Fields        []Field
	DerivedFields []DerivedField
	Bricks        map[string]Brick
	Equations     []string
	RHS           []RHS
}

// NewModel returns a new model
func NewModel() Model {
	return Model{
		Bricks: make(map[string]Brick),
	}
}

// AddField adds a field to the model
func (m *Model) AddField(f Field) {
	m.Fields = append(m.Fields, f)
	m.Bricks[f.Name] = &f
}

// AddScalar adds a scalar to the model
func (m *Model) AddScalar(s Scalar) {
	m.Bricks[s.Name] = &s
}

// AddEquation adds equations to the model
func (m *Model) AddEquation(eq string) {
	eq = strings.Replace(eq, " ", "", -1)
	m.Equations = append(m.Equations, eq)
	m.UpdateDerivedFields(eq)
}

// UpdateDerivedFields update fields that needs to be handle with FFT (required for non-linear equations)
func (m *Model) UpdateDerivedFields(eq string) {
	rhs := strings.Split(eq, "=")[1]
	splitted := strings.Split(rhs, "+")
	field := fieldNameFromLeibniz(strings.Split(eq, "=")[0])
	fieldNames := make([]string, len(m.Fields))
	for i := range m.Fields {
		fieldNames[i] = m.Fields[i].Name
	}

	for i := range splitted {
		newFields := GetNonLinearFieldExpressions(splitted[i], field, fieldNames)

		if newFields != "" && !m.IsFieldName(newFields) {
			dField := DerivedField{
				Data: make([]complex128, len(m.Fields[0].Data)),
				Name: newFields,
				Calc: DerivedFieldCalcFromDesc(newFields, m.Fields),
			}
			m.DerivedFields = append(m.DerivedFields, dField)
			m.Bricks[newFields] = &dField
		}
	}
}

// AllFieldNames returns all field names (including derived fields)
func (m *Model) AllFieldNames() []string {
	names := make([]string, len(m.Fields)+len(m.DerivedFields))
	for i, f := range m.Fields {
		names[i] = f.Name
	}

	for i, f := range m.DerivedFields {
		names[len(m.Fields)+i] = f.Name
	}
	return names
}

// IsFieldName checks if the passed name is a field name
func (m *Model) IsFieldName(name string) bool {
	for _, f := range m.Fields {
		if f.Name == name {
			return true
		}
	}

	for _, f := range m.DerivedFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// IsBrickName returns true if a brick with the passed name exists
func (m *Model) IsBrickName(name string) bool {
	for n := range m.Bricks {
		if n == name {
			return true
		}
	}
	return false
}

// SyncDerivedFields updates all derived fields
func (m *Model) SyncDerivedFields() {
	for _, f := range m.DerivedFields {
		f.Update()
	}
}

// Init prepares the model
func (m *Model) Init() {
	m.RHS = m.RHS[:0]
	for _, eq := range m.Equations {
		m.RHS = append(m.RHS, Build(eq, m))
	}
	m.SyncDerivedFields()
}

// AllVariableNames return all known variable names
func (m *Model) AllVariableNames() []string {
	names := make([]string, len(m.Bricks))
	idx := 0
	for n := range m.Bricks {
		names[idx] = n
		idx++
	}
	return names
}

// GetRHS evaluates the right hand side of one of the equations
func (m *Model) GetRHS(fieldNo int, freq Frequency, t float64) []complex128 {
	data := make([]complex128, len(m.Fields[fieldNo].Data))
	tmp := make([]complex128, len(m.Fields[fieldNo].Data))
	for _, f := range m.RHS[fieldNo].Terms {
		f(freq, t, tmp)
		ElemwiseAdd(data, tmp)
	}
	return data
}

// GetDenum evaluates the denuminator
func (m *Model) GetDenum(fieldNo int, freq Frequency, t float64) []complex128 {
	data := make([]complex128, len(m.Fields[fieldNo].Data))
	tmp := make([]complex128, len(m.Fields[fieldNo].Data))
	for _, f := range m.RHS[fieldNo].Denum {
		f(freq, t, tmp)
		ElemwiseAdd(data, tmp)
	}
	return data
}
