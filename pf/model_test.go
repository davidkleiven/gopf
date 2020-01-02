package pf

import (
	"math"
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
