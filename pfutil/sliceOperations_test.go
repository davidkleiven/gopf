package pfutil

import (
	"math"
	"testing"
)

func TestCmplxEqualApprox(t *testing.T) {
	for i, test := range []struct {
		A      []complex128
		B      []complex128
		Expect bool
	}{
		{
			A:      []complex128{complex(1.0, 0.1), complex(2.0, -0.3)},
			B:      []complex128{complex(1.0, 0.1), complex(2.0, -0.3)},
			Expect: true,
		},
		{
			A:      []complex128{complex(1.0, 0.2), complex(2.0, -0.3)},
			B:      []complex128{complex(1.0, 0.1), complex(2.0, -0.3)},
			Expect: false,
		},
	} {
		res := CmplxEqualApprox(test.A, test.B, 1e-10)
		if res != test.Expect {
			t.Errorf("Test #%d: Expected %v got %v\n", i, test.Expect, res)
		}
	}
}

func TestElemwiseAdd(t *testing.T) {
	dst := []complex128{complex(1.0, 1.0),
		complex(2.0, 0.0),
		complex(3.0, 0.1)}
	add := []complex128{complex(0.0, 1.0),
		complex(2.0, 1.0),
		complex(2.0, -0.1)}
	expect := []complex128{complex(1.0, 2.0),
		complex(4.0, 1.0),
		complex(5.0, 0.0)}
	ElemwiseAdd(dst, add)
	if !CmplxEqualApprox(dst, expect, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, dst)
	}
}

func TestElemwiseMul(t *testing.T) {
	dst := []complex128{
		complex(1.0, 0.2),
		complex(2.0, 1.0),
	}

	factor := []complex128{
		complex(2.0, 1.0),
		complex(-1.0, -3.0),
	}
	ElemwiseMul(dst, factor)
	expect := []complex128{
		complex(2.0-0.2, 1.0+0.4),
		complex(-2.0+3.0, -1.0-6.0),
	}

	if !CmplxEqualApprox(dst, expect, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, dst)
	}
}

func TestDivRealScalar(t *testing.T) {
	data := []complex128{
		complex(1.0, 2.0), complex(-1.0, 0.3),
	}
	factor := 3.0
	DivRealScalar(data, factor)
	expect := []complex128{
		complex(1.0/factor, 2.0/factor), complex(-1.0/factor, 0.3/factor),
	}

	if !CmplxEqualApprox(data, expect, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, data)
	}
}

func TestProdInt(t *testing.T) {
	data := []int{1, 2, 5, 4}
	expect := 40
	if ProdInt(data) != expect {
		t.Errorf("Expected %d got %d\n", expect, ProdInt(data))
	}
}

func TestDot(t *testing.T) {
	a := []float64{1.0, 2.0}
	b := []float64{-2.0, 3.0}
	expect := 4.0
	got := Dot(a, b)
	if math.Abs(expect-got) > 1e-10 {
		t.Errorf("Expected %f got %f\n", expect, got)
	}
}

func TestMaxReal(t *testing.T) {
	data := []complex128{
		complex(1.0, -2.0),
		complex(0.2, 4.0),
	}
	max := MaxReal(data)
	if math.Abs(max-1.0) > 1e-10 {
		t.Errorf("Expectd 1.0 got %f", max)
	}
}

func TestMinReal(t *testing.T) {
	data := []complex128{
		complex(1.0, -2.0),
		complex(0.2, 4.0),
	}
	max := MinReal(data)
	if math.Abs(max-0.2) > 1e-10 {
		t.Errorf("Expectd 1.0 got %f", max)
	}
}

func TestClear(t *testing.T) {
	data := []complex128{
		complex(1.0, 0.2), complex(2.0, 4.0),
	}
	expect := make([]complex128, 2)
	Clear(data)
	if !CmplxEqualApprox(data, expect, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, data)
	}
}
