package pf

import (
	"math"
	"math/cmplx"
	"sort"
	"testing"
)

func Freq(i int) []float64 {
	return []float64{float64(i), float64(i)}
}

func TestTermDiffusion(t *testing.T) {
	m := NewModel()
	conc := NewField("conc", 2, []complex128{complex(1.0, 0.0), complex(2.0, 0.0)})
	m.AddField(conc)
	m.AddEquation("dconc/dt = LAP conc")
	m.Init()

	if len(m.RHS[0].Terms) != 0 {
		t.Errorf("Unexpected number of terms")
	}

	if len(m.RHS[0].Denum) != 1 {
		t.Errorf("Unexpected number of bilinear terms")
	}

	// Evaluate RHS
	twoPiSq := math.Pow(2.0*math.Pi, 2.0)
	values := make([]complex128, len(conc.Data))
	m.RHS[0].Denum[0](Freq, 0.0, values)
	expect := []complex128{complex(0.0, 0.0), complex(-2.0*twoPiSq, 0.0)}

	if !CmplxEqualApprox(expect, values, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, values)
	}
}

func TestReactionDiffusion(t *testing.T) {
	m := NewModel()
	concA := NewField("concA", 2, []complex128{complex(1.0, 0.0), complex(2.0, 0.0)})
	concB := NewField("concB", 2, []complex128{complex(3.0, 0.0), complex(5.0, 0.0)})
	concC := NewField("concC", 2, []complex128{complex(-1.0, 0.0), complex(1.0, 0.0)})
	m1 := NewScalar("m1", complex(-1.0, 0.0))
	kf := NewScalar("kf", complex(2.0, 0.0))
	kr := NewScalar("kr", complex(0.2, 0.0))
	m.AddField(concA)
	m.AddField(concB)
	m.AddField(concC)
	m.AddScalar(m1)
	m.AddScalar(kf)
	m.AddScalar(kr)

	// Diffusion + the reaction 2A + 3B <> C
	m.AddEquation("dconcA/dt = LAP concA + kf*m1*concA^2*concB^3 + kr*concC")
	m.AddEquation("dconcB/dt = LAP concB + kf*m1*concA^2*concB^3 + kr*concC")
	m.AddEquation("dconcC/dt = LAP concC + kr*m1*concC + kf*concA^2*concB^3")
	m.Init()

	expectedFields := []string{"concA", "concB", "concC", "concA^2*concB^3"}
	res := m.AllFieldNames()
	sort.Strings(res)
	sort.Strings(expectedFields)

	if len(res) != len(expectedFields) {
		t.Errorf("Wrong number of fields\nExpected\n%v\nGot\n%v\n", expectedFields, res)
	} else {
		for i := range res {
			if res[i] != expectedFields[i] {
				t.Errorf("Wrong fields. Expected %s got %s", expectedFields[i], res[i])
			}
		}
	}

	if len(m.RHS) != 3 {
		t.Errorf("Expected 3 equations")
	}

	for i, test := range []struct {
		numTerms int
		numDenum int
	}{
		{
			numTerms: 2,
			numDenum: 1,
		},
		{
			numTerms: 2,
			numDenum: 1,
		},
		{
			numTerms: 1,
			numDenum: 2,
		},
	} {
		if len(m.RHS[i].Terms) != test.numTerms {
			t.Errorf("Test #%d: Wrong number of terms. Expected %d got %d", i, len(m.RHS[i].Terms), test.numTerms)
		}

		if len(m.RHS[i].Denum) != test.numDenum {
			t.Errorf("Test #%d: Wrong number of denums. Expected %d got %d", i, len(m.RHS[i].Denum), test.numTerms)
		}
	}
}

// LapDensitySquared is a struct that is used to represent the term
// nabla^2 density^2, where u is a function
type LapDensitySquared struct {
	numCallConstruct int
}

// Construct returns the function needed to evaluate the fourier transformed
// version of the term
func (l *LapDensitySquared) Construct(bricks map[string]Brick) Term {
	l.numCallConstruct++
	lap := LaplacianN{Power: 1}
	return func(freq Frequency, t float64, field []complex128) []complex128 {
		for i := range field {
			field[i] = bricks["density^2"].Get(i)
		}
		lap.Eval(freq, field)
		return field
	}
}

// OnStepFinished does nothing for this term
func (l *LapDensitySquared) OnStepFinished(t float64, bricks map[string]Brick) {}

// GetUsquared returns a function that calculates u-squared
func GetUsquared(fields []Field) DerivedFieldCalc {
	for _, f := range fields {
		if f.Name == "density" {
			return func(out []complex128) {
				for i := range f.Data {
					out[i] = cmplx.Pow(f.Data[i], 2)
				}
			}
		}
	}
	panic("No field called density!")
}

func TestUserDefinedTerms(t *testing.T) {
	N := 64
	model := NewModel()
	field := NewField("density", N*N, nil)
	model.AddField(field)

	// Initialize the user defined term
	var lapUsq LapDensitySquared
	dField := DerivedField{
		Name: "density^2",
		Calc: GetUsquared(model.Fields),
		Data: make([]complex128, N*N),
	}
	model.RegisterUserDefinedTerm("LAP_DENSITY_SQUARED", &lapUsq, []DerivedField{dField})
	model.AddEquation("ddensity/dt = LAP_DENSITY_SQUARED")
	model.Init()

	// Check status
	if len(model.Fields) != 1 {
		t.Errorf("Unexpected number of fields. Expected 1 got %d", len(model.Fields))
	}

	if model.Fields[0].Name != "density" {
		t.Errorf("Expected density got %s", model.Fields[0].Name)
	}

	if len(model.DerivedFields) != 1 {
		t.Errorf("Expected 1 derived field. Got %d", len(model.DerivedFields))
	}

	if model.DerivedFields[0].Name != "density^2" {
		t.Errorf("Expected first derived field to be called density^2. Got %s", model.DerivedFields[0].Name)
	}

	if len(model.UserDef) != 1 {
		t.Errorf("Expected 1 user defined field. Got %d", len(model.UserDef))
	}

	if lapUsq.numCallConstruct != 1 {
		t.Errorf("Expected 1 call to Construct. got %d", lapUsq.numCallConstruct)
	}

	if len(model.RHS[0].Terms) != 1 {
		t.Errorf("Expected 1 term in the right hand side. Got %d", len(model.RHS[0].Terms))
	}

	if len(model.RHS[0].Denum) != 0 {
		t.Errorf("Expected 0 terms in the denuminator. Got %d", len(model.RHS[0].Denum))
	}
}

func TestFunction(t *testing.T) {
	model := NewModel()

	N := 8
	f := NewField("myfield", N, nil)
	for i := 0; i < N; i++ {
		f.Data[i] = complex(float64(i), 0.0)
	}
	model.AddField(f)

	model.RegisterFunction("myfunc", func(i int, bricks map[string]Brick) complex128 {
		return bricks["myfield"].Get(i)
	})

	function := model.UserDef["myfunc"].Construct(model.Bricks)
	model.SyncDerivedFields()
	res := make([]complex128, N)
	function(func(i int) []float64 { return []float64{} }, 0.0, res)

	tol := 1e-10
	for i := 0; i < N; i++ {
		if math.Abs(real(res[i])-float64(i)) > tol {
			t.Errorf("Expected %d got %v", i, res[i])
		}
	}
}
