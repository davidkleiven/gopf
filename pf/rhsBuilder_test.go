package pf

import (
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
		values []complex128
		expect []complex128
	}{
		{
			expr:   "resistance*current^2",
			expect: []complex128{complex(8.0, 0.0)},
		},
		{
			expr:   "voltage^2",
			expect: []complex128{complex(16.0, 0.0)},
		},
	} {
		term := ConcreteTerm(test.expr, &model)
		res := make([]complex128, len(test.expect))
		got := term(Freq, 0.0, res)

		if !CmplxEqualApprox(got, test.expect, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, test.expect, got)
		}
	}
}
