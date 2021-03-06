package pf

import (
	"fmt"
	"regexp"
	"strings"
	"log"

	"github.com/davidkleiven/gopf/pfutil"
)

// RHSModifier is a function type used to modify the right hand side prior to adding it
type RHSModifier func(data []complex128)

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

// Copy returns a new field that is a deep copy of the current
func (f Field) Copy() Field {
	field := NewField(f.Name, len(f.Data), nil)
	copy(field.Data, f.Data)
	return field
}

// SaveReal stores the real part as a raw binary file with big endian
func (f Field) SaveReal(fname string) {
	realPart := make([]float64, len(f.Data))
	for j := range f.Data {
		realPart[j] = real(f.Data[j])
	}
	SaveFloat64(fname, realPart)
}

// NewField initializes a new field
func NewField(name string, N int, data []complex128) Field {
	panicOnPrefixInName(name)
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
	panicOnPrefixInName(name)
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

type eqModifier struct {
	EqNo        int
	RHSModifier RHSModifier
}

// Model is a type used to represent a general set of equations
type Model struct {
	Fields        []Field
	DerivedFields []DerivedField
	Bricks        map[string]Brick
	ImplicitTerms map[string]PureTerm
	ExplicitTerms map[string]PureTerm
	MixedTerms    map[string]MixedTerm
	Equations     []string
	RHS           []RHS
	AllSources    []Sources
	RHSModifiers  []eqModifier
}

// NewModel returns a new model
func NewModel() Model {
	return Model{
		Bricks:        make(map[string]Brick),
		ImplicitTerms: make(map[string]PureTerm),
		ExplicitTerms: make(map[string]PureTerm),
		MixedTerms:    make(map[string]MixedTerm),
		RHSModifiers:  []eqModifier{},
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

// AddSource adds a source to the equation
func (m *Model) AddSource(eqNo int, s Source) {
	m.AllSources[eqNo] = append(m.AllSources[eqNo], s)
}

// AddEquation adds equations to the model
func (m *Model) AddEquation(eq string) {
	eq = strings.Replace(eq, " ", "", -1)
	m.Equations = append(m.Equations, eq)
	m.UpdateDerivedFields(eq)
	m.AllSources = append(m.AllSources, make(Sources, 0))
}

// UpdateDerivedFields update fields that needs to be handle with FFT (required for non-linear equations)
func (m *Model) UpdateDerivedFields(eq string) {
	rhs := strings.Split(eq, "=")[1]
	splittedDelim := SplitOnMany(rhs, []string{"+", "-"})
	splitted := make([]string, len(splittedDelim))
	for i := range splittedDelim {
		splitted[i] = splittedDelim[i].SubString
	}

	field := fieldNameFromLeibniz(strings.Split(eq, "=")[0])
	fieldNames := make([]string, len(m.Fields))
	for i := range m.Fields {
		fieldNames[i] = m.Fields[i].Name
	}

	for i := range splitted {
		if m.IsUserDefinedTerm(splitted[i]) {
			continue
		}
		newFields := SortFactors(GetNonLinearFieldExpressions(splitted[i], field, fieldNames))

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

	// Validate all implicit terms
	for k, v := range m.ImplicitTerms {
		if !isImplicit(v, m.Bricks, m.NumNodes(), func(i int) []float64 {
			return []float64{0.4, 0.4}
		}) {
			msg := fmt.Sprintf("Model: Term %s is not implicit (e.g. it varies when the fields are varied)", k)
			panic(msg)
		}
	}
}

// NumNodes returns the number of nodes in the simulation cell. It panics if no
// fields are added
func (m *Model) NumNodes() int {
	if len(m.Fields) == 0 {
		panic("Model: No fields added")
	}
	return len(m.Fields[0].Data)
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
		pfutil.ElemwiseAdd(data, tmp)
	}

	for _, s := range m.AllSources[fieldNo] {
		s.Eval(freq, t, tmp)
		pfutil.ElemwiseAdd(data, tmp)
	}

	// Apply eventual modifiers
	for i := range m.RHSModifiers {
		if m.RHSModifiers[i].EqNo == fieldNo {
			m.RHSModifiers[i].RHSModifier(data)
		}
	}
	return data
}

// GetDenum evaluates the denuminator
func (m *Model) GetDenum(fieldNo int, freq Frequency, t float64) []complex128 {
	data := make([]complex128, len(m.Fields[fieldNo].Data))
	tmp := make([]complex128, len(m.Fields[fieldNo].Data))
	for _, f := range m.RHS[fieldNo].Denum {
		f(freq, t, tmp)
		pfutil.ElemwiseAdd(data, tmp)
	}
	return data
}

const (
	implicitTerm = iota
	explicitTerm
)

// registerTerm defines a new pure term (linear og non linear)
func (m *Model) registerTerm(name string, t PureTerm, dFields []DerivedField, termType int) {
	panicOnPrefixInName(name)
	switch termType {
	case implicitTerm:
		m.ImplicitTerms[name] = t
	case explicitTerm:
		m.ExplicitTerms[name] = t
	default:
		panic("Has to be either linear or a non-linear term")
	}
	m.registerDerivedFields(dFields)
}

// RegisterDerivedFields adds a new set of derived fields to the model
func (m *Model) registerDerivedFields(dFields []DerivedField) {
	if dFields != nil {
		for _, f := range dFields {
			if !m.IsFieldName(f.Name) {
				m.DerivedFields = append(m.DerivedFields, f)
				m.Bricks[f.Name] = &f
			}
		}
	}
}

// RegisterImplicitTerm can be used to register terms if the form
// f({otherfields])*field
func (m *Model) RegisterImplicitTerm(name string, t PureTerm, dFields []DerivedField) {
	m.registerTerm(name, t, dFields, implicitTerm)
}

// RegisterExplicitTerm defines a new term. To add the term to an equation add the
// name as one of the terms.
//
// Example:
// If there is a user defined term called LINEAR_ELASTICITY, and we have a PDE where
// the term is present on the right hand side, the equation would look like
// dx/dt = LINEAR_ELASTICITY
// where x is the name of the field. The additional derived fields (which are fields that are
// contructed from the original fields) is specified via dFields
func (m *Model) RegisterExplicitTerm(name string, t PureTerm, dFields []DerivedField) {
	m.registerTerm(name, t, dFields, explicitTerm)
}

// RegisterMixedTerm is used to register terms that contains a linear part and a
// non-linear part. The linear part will be treated implicitly during time evolution,
// while the non-linear part is treated explicitly
func (m *Model) RegisterMixedTerm(name string, t MixedTerm, dFields []DerivedField) {
	panicOnPrefixInName(name)
	m.MixedTerms[name] = t
	m.registerDerivedFields(dFields)
}

// IsImplicitTerm checks if the given term is a linear term
func (m *Model) IsImplicitTerm(desc string) bool {
	_, ok := m.ImplicitTerms[desc]
	return ok
}

// IsExplicitTerm returns true if the passed string is a non-linear term
func (m *Model) IsExplicitTerm(desc string) bool {
	_, ok := m.ExplicitTerms[desc]
	return ok
}

// IsMixedTerm returns true if the passed string is a mixed term
func (m *Model) IsMixedTerm(desc string) bool {
	_, ok := m.MixedTerms[desc]
	return ok
}

// IsUserDefinedTerm returns true if desc matches either one of the linear terms,
// non-linear terms or mixed terms
func (m *Model) IsUserDefinedTerm(desc string) bool {
	return m.IsImplicitTerm(desc) || m.IsExplicitTerm(desc) || m.IsMixedTerm(desc)
}

// RegisterFunction registers a function that may be used in the equations
func (m *Model) RegisterFunction(name string, F GenericFunction) {
	panicOnPrefixInName(name)
	dField := DerivedField{
		Data: make([]complex128, len(m.Fields[0].Data)),
		Name: name,
		Calc: func(out []complex128) {
			for i := range out {
				out[i] = F(i, m.Bricks)
			}
		},
	}
	m.RegisterDerivedField(dField)
}

// RegisterDerivedField registers a new derived field
func (m *Model) RegisterDerivedField(d DerivedField) {
	m.DerivedFields = append(m.DerivedFields, d)
	m.Bricks[d.Name] = &d
}

// Summarize prints a summary of the model
func (m *Model) Summarize() {
	if len(m.RHS) != len(m.Equations) {
		fmt.Printf("Model not initialized - summary not available\n")
		return
	}
	fmt.Printf("=========================================================================================\n")
	fmt.Printf("                                    MODEL SUMMARY                                        \n")
	fmt.Printf("=========================================================================================\n")
	fmt.Printf("NE - Number of expclit terms in time stepping\n")
	fmt.Printf("NI - Number of implicit terms in time stepping\n")
	fmt.Printf("-----------------------------------------------------------------------------------------\n")
	fmt.Printf("| Eq |                         String representation                          | NE | NI |\n")
	fmt.Printf("-----------------------------------------------------------------------------------------\n")
	for i := range m.Equations {
		fmt.Printf("| %2d | %-70s | %2d | %2d |\n", i, m.Equations[i], len(m.RHS[i].Terms), len(m.RHS[i].Denum))
	}
	fmt.Printf("=========================================================================================\n")
}

// EqNumber returns the equation number corresponding to the passed field name
func (m *Model) EqNumber(fieldName string) int {
	re := regexp.MustCompile("d(.*?)/dt")
	for i, eq := range m.Equations {
		match := re.FindStringSubmatch(eq) // Full match is first element and capture group is the second
		if len(match) >= 2 {
			if match[1] == fieldName {
				return i
			}
		}
	}
	msg := fmt.Sprintf("EqNumber: Could not find an equation for field %s\n", fieldName)
	panic(msg)
}

// RegisterRHSModifier adds a modifier that is applied to the fourier transformed right hand
// side of the equation. The modifier will be applied to the right hand side of the equation
// passed. If several modifiers are added to the same equation, the one that is added first
// will be applied first
//
// Simple example case:
// If we want to convert the units of the right from let's milli joules to joules, we should
// multiply the current right hand side by a factor 1e-3. We therefore supply the following
// modifier
//
// func (data []complex128) {
//	for i := range data {
//		data[i] *= 1e-3
//	}
//}
//
// The frequency corresponding to each node can be obtained from the frequency method
// of the fourier transformer used (see for example FFTWWrapper)
func (m *Model) RegisterRHSModifier(eqNumber int, modifier func(data []complex128)) {
	m.RHSModifiers = append(m.RHSModifiers, eqModifier{
		EqNo:        eqNumber,
		RHSModifier: modifier,
	})
}

func panicOnPrefixInName(name string) {
	if len(getKnownPrefixes(name)) > 0 {
		prefixes := knownPrefixes()
		log.Fatalf("The words %v are reserved. Do not include them in your variable names\n", prefixes)
	}
}