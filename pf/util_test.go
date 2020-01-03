package pf

import (
	"math"
	"testing"
)

func TestGetNonLinearFieldExp(t *testing.T) {
	fieldNames := []string{"conc1", "conc2", "eta1", "eta2"}
	for i, test := range []struct {
		expr   string
		field  string
		expect string
	}{
		{
			expr:   "conc1^2*eta1*factor",
			field:  "conc1",
			expect: "conc1^2*eta1",
		},
		{
			expr:   "conc1^2*eta1*factor",
			field:  "eta1",
			expect: "conc1^2",
		},
		{
			expr:   "conc2*conc1",
			field:  "conc2",
			expect: "conc1",
		},
		{
			expr:   "LAPconc2^2*eta2^3",
			field:  "conc2",
			expect: "conc2^2*eta2^3",
		},
	} {
		got := GetNonLinearFieldExpressions(test.expr, test.field, fieldNames)
		if got != test.expect {
			t.Errorf("Test #%d: Expected %s Got %s", i, test.expect, got)
		}
	}
}

func TestDerivedCalcFromDesc(t *testing.T) {
	fields := []Field{
		NewField("conc1", 2, []complex128{complex(1.0, 0.0), complex(2.0, 0.0)}),
		NewField("conc2", 2, []complex128{complex(3.0, 0.0), complex(4.0, 0.0)}),
		NewField("conc3", 2, []complex128{complex(5.0, 0.0), complex(6.0, 0.0)}),
	}

	for i, test := range []struct {
		desc   string
		expect []complex128
	}{
		{
			desc:   "conc1^2*conc2",
			expect: []complex128{complex(3.0, 0.0), complex(16.0, 0.0)},
		},
		{
			desc:   "conc3^2*conc2",
			expect: []complex128{complex(75.0, 0.0), complex(144.0, 0.0)},
		},
	} {
		calc := DerivedFieldCalcFromDesc(test.desc, fields)
		array := make([]complex128, 2)
		calc(array)
		if !CmplxEqualApprox(test.expect, array, 1e-10) {
			t.Errorf("Test #%d: Expected\n%v\nGot\n%v\n", i, test.expect, array)
		}
	}
}

func TestGetPower(t *testing.T) {
	for i, test := range []struct {
		expr   string
		expect float64
	}{
		{
			expr:   "conc1^2",
			expect: 2.0,
		},
		{
			expr:   "conc1",
			expect: 1.0,
		},
		{
			expr:   "conc4^-4.5",
			expect: -4.5,
		},
	} {
		got := GetPower(test.expr)
		if math.Abs(got-test.expect) > 1e-10 {
			t.Errorf("Test #%d: Expected %f got %f", i, got, test.expect)
		}
	}
}

func TestGetFieldName(t *testing.T) {
	names := []string{"conc1", "conc2", "conc3", "conc1^2*conc2", "conc3^3", "conc2^4*conc1^2", "conc1^2"}
	for i, test := range []struct {
		expr   string
		expect string
	}{
		{
			expr:   "conc1^2*conc2*otherstuff",
			expect: "conc1^2*conc2",
		},
		{
			expr:   "*randomstuff*conc1^2*otherstuff",
			expect: "conc1^2",
		},
	} {
		got := GetFieldName(test.expr, names)
		if got != test.expect {
			t.Errorf("Test #%d: Expected %s got %s", i, test.expect, got)
		}
	}
}
