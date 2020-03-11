package pf

import (
	"math"
	"testing"
)

func TestNameFromLeibniz(t *testing.T) {
	for i, test := range []struct {
		leibniz string
		expect  string
	}{
		{
			leibniz: "dc/dt",
			expect:  "c",
		},
		{
			leibniz: "dkappa/dt",
			expect:  "kappa",
		},
	} {
		got := fieldNameFromLeibniz(test.leibniz)
		if got != test.expect {
			t.Errorf("Test #%d: Expeted %s got %s", i, test.expect, got)
		}
	}

	// Test panics
	shouldPanic := []string{"dc", "ac/dt", "dc/dq"}
	for i := range shouldPanic {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Test #%d: Did  not panic", i)
				}
			}()
			fieldNameFromLeibniz(shouldPanic[i])
		}()
	}
}

func TestIsBilinear(t *testing.T) {
	for i, test := range []struct {
		field  string
		expr   string
		expect bool
	}{
		{
			field:  "c",
			expr:   "2*c",
			expect: true,
		},
		{
			field:  "conc",
			expr:   "conc^2",
			expect: false,
		},
		{
			field:  "c",
			expr:   "c*n*r",
			expect: true,
		},
		{
			field:  "voltage",
			expr:   "voltage^1.62",
			expect: false,
		},
		{
			field:  "voltage",
			expr:   "current*voltage^1.0",
			expect: true,
		},
		{
			field:  "current",
			expr:   "P*current^-2",
			expect: false,
		},
	} {
		got := isBilinear(test.expr, test.field)
		if test.expect != got {
			t.Errorf("Test #%d: expected %v got %v", i, test.expect, got)
		}
	}
}

func TestConcreteTerm(t *testing.T) {
	model := NewModel()
	resistance := NewScalar("resistance", complex(2.0, 0.0))
	current := NewField("current", 1, []complex128{complex(2.0, 0.0)})
	voltage := NewField("voltage", 1, []complex128{complex(-4.0, 0.0)})
	magneticPot := NewField("magnetic", 1, []complex128{complex(1.5, 0.0)})

	model.AddField(current)
	model.AddField(voltage)
	model.AddField(magneticPot)
	model.AddScalar(resistance)

	// Add equations for all fields
	model.AddEquation("dcurrent/dt = voltage^2")
	model.AddEquation("dvoltage/dt = resistance*current^2")
	model.AddEquation("dmagnetic/dt = current*magnetic^3")
	model.SyncDerivedFields()

	Freq := func(i int) []float64 {
		return []float64{1.0, 1.0}
	}

	for i, test := range []struct {
		expr   string
		sign   string
		values []complex128
		expect []complex128
	}{
		{
			expr:   "resistance*current^2",
			expect: []complex128{complex(8.0, 0.0)},
			sign:   "+",
		},
		{
			expr:   "voltage^2",
			expect: []complex128{complex(16.0, 0.0)},
			sign:   "+",
		},
		{
			expr:   "resistance*current^2",
			expect: []complex128{complex(-8.0, 0.0)},
			sign:   "-",
		},
		{
			expr:   "voltage^2",
			expect: []complex128{complex(-16.0, 0.0)},
			sign:   "-",
		},
	} {
		substring := SubStringDelimiter{
			SubString:           test.expr,
			PreceedingDelimiter: test.sign,
		}
		term := ConcreteTerm(substring, &model)
		got := make([]complex128, len(test.expect))
		term(Freq, 0.0, got)

		if !CmplxEqualApprox(got, test.expect, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, test.expect, got)
		}
	}
}

func TestPanicOnUnknownName(t *testing.T) {
	model := NewModel()
	field := NewField("conc", 8, nil)
	model.AddField(field)

	for i, test := range []struct {
		expr        string
		shouldPanic bool
	}{
		{
			expr:        "LAP conc",
			shouldPanic: false,
		},
		{
			expr:        "conc",
			shouldPanic: false,
		},
		{
			expr:        "m1*conc",
			shouldPanic: true,
		},
		{
			expr:        "m1*LAP conc",
			shouldPanic: true,
		},
		{
			expr:        "LAP otherField",
			shouldPanic: true,
		},
	} {
		func() {
			defer func() {
				if test.shouldPanic {
					if recover() == nil {
						t.Errorf("Test %d should have panicked\n", i)
					}
				} else {
					if recover() != nil {
						t.Errorf("Unexpected panic in test %d\n", i)
					}
				}
			}()
			substring := SubStringDelimiter{
				SubString: test.expr,
			}
			ConcreteTerm(substring, &model)
		}()
	}
}

func TestLapUserDefined(t *testing.T) {
	function := func(i int, bricks map[string]Brick) complex128 {
		return bricks["conc"].Get(i)
	}

	model := NewModel()
	N := 16
	field := NewField("conc", N*N, nil)

	// Set fourier transformed field values
	for i := range field.Data {
		field.Data[i] = complex(0.1*float64(i), 0.0)
	}

	model.AddField(field)
	model.RegisterFunction("myfunc", function)
	model.AddEquation("dconc/dt = LAP myfunc")
	model.Init()
	terms := model.RHS[0].Terms

	if len(terms) != 1 {
		t.Errorf("Expected 1 term got %d\n", len(terms))
	}

	freq := NewFFTW([]int{N, N}).Freq

	termEval := make([]complex128, N*N)
	terms[0](freq, 0.0, termEval)

	for i := range termEval {
		f := freq(i)
		fRadSq := f[0]*f[0] + f[1]*f[1]
		expect := -4.0 * math.Pi * math.Pi * fRadSq * real(field.Data[i])
		re := real(termEval[i])

		if math.Abs(expect-re) > 1e-10 {
			t.Errorf("Expected %f got %f\n", expect, re)
		}
	}
}
