package pf

import (
	"math"
	"sort"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
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
			expect: "conc1^2*eta1",
		},
		{
			expr:   "conc2*conc1",
			field:  "conc2",
			expect: "conc1*conc2",
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
		if !pfutil.CmplxEqualApprox(test.expect, array, 1e-10) {
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

type DummyFilter struct{}

func (df *DummyFilter) Eval(x float64) float64 {
	return 0.5
}

func TestApplyModalFilter(t *testing.T) {
	data := make([]complex128, 10)
	for i := range data {
		data[i] = complex(float64(i), 0.0)
	}

	ApplyModalFilter(&DummyFilter{}, func(i int) []float64 { return []float64{0.0} }, data)

	tol := 1e-10
	for i := range data {
		re := real(data[i])
		im := imag(data[i])
		if math.Abs(re-0.5*float64(i)) > tol || math.Abs(im) > 0.0 {
			t.Errorf("Expected 0.5 got (%f, %f)\n", re, im)
		}
	}
}

func TestSplitOnMany(t *testing.T) {
	for i, test := range []struct {
		Value  string
		Expect []string
		Delims []string
	}{
		{
			Value:  "a+b",
			Delims: []string{"+"},
			Expect: []string{"a", "b"},
		},
		{
			Value:  "a+b+cd",
			Delims: []string{"+"},
			Expect: []string{"a", "b", "cd"},
		},
		{
			Value:  "a+b-cd",
			Delims: []string{"+", "-"},
			Expect: []string{"a", "b", "cd"},
		},
		{
			Value:  "cdb-the+two",
			Delims: []string{"+", "-", "7"},
			Expect: []string{"cdb", "the", "two"},
		},
	} {
		res := SplitOnMany(test.Value, test.Delims)
		substr := make([]string, len(res))
		for i := range res {
			substr[i] = res[i].SubString
		}
		sort.Strings(substr)

		for j := range substr {
			if substr[j] != test.Expect[j] {
				t.Errorf("Test #%d: Expected %s got %s\n", i, test.Expect[j], substr[j])
			}
		}
	}
}

func TestUniqueFreqIterator(t *testing.T) {
	N := 16
	for i, test := range []struct {
		DomainSize       []int
		ExpectNum        int
		ExpectNumNyquist int
	}{
		{
			DomainSize: []int{N, N},

			// ExpectNum = N*N/2 + 2 because we have inversion symmetry, thus
			// we can restrict ourselves to the case when fy >= 0.0. If a grid
			// has N*N grid points, there should N*N independent fourier coefficients.
			// Since we have both real and imaginary parts, the total number of "unknowns"
			// is 2*(N*N/2 + 2) = N*N + 4. Furthermore, we know that the fourier amplitudes
			// at the frequencies (0, 0), (0, 0.5), (0.5, 0) and (0.5, 0.5) are real. Thus,
			// the 4 imaginary parts of these amplitudes are not unknown. Consequently, there
			// are only N*N independent "unknowns"
			ExpectNum:        N*N/2 + 2,
			ExpectNumNyquist: 1,
		},
		{
			DomainSize: []int{N, N, N},

			// Same reasoning as in the 2D case. But frequencies
			// (0, 0, 0), (0, 0, 0.5), (0, 0.5, 0), (0.5, 0, 0), (0.5, 0.5, 0)
			// (0.5, 0, 0.5), (0, 0.5, 0.5) and (0.5, 0.5, 0.5) are real.
			// This removes 8 imaginary parts
			ExpectNum:        N*N*N/2 + 4,
			ExpectNumNyquist: 1,
		},
	} {
		ft := NewFFTW(test.DomainSize)
		num := 0
		iterator := UniqueFreqIterator{
			Freq: ft.Freq,
			End:  pfutil.ProdInt(test.DomainSize),
		}
		zeroFound := false
		numNyquist := 0
		for i := iterator.Next(); i != -1; i = iterator.Next() {
			num++
			f := ft.Freq(i)

			if !zeroFound {
				zeroFound = true
				for j := range f {
					if math.Abs(f[j]) > 1e-10 {
						zeroFound = false
						break
					}
				}
			}

			allCloseToHalf := true
			for j := range f {
				if math.Abs(math.Abs(f[j])-0.5) > 1e-10 {
					allCloseToHalf = false
					break
				}
			}

			if allCloseToHalf {
				numNyquist++
			}
		}

		if num != test.ExpectNum {
			t.Errorf("Test #%d: Expected %d items got %d\n", i, test.ExpectNum, num)
		}

		if !zeroFound {
			t.Errorf("Test #%d: Zero frequency not found\n", i)
		}

		if numNyquist != test.ExpectNumNyquist {
			t.Errorf("Test #%d: Expected %d nyquist frequencies got %d\n", i, test.ExpectNumNyquist, numNyquist)
		}
	}
}

func TestRealAmplitudeIterator(t *testing.T) {
	for i, test := range []struct {
		DomainSize []int
		ExpectNum  int
	}{
		{
			DomainSize: []int{8, 8},
			ExpectNum:  4,
		},
	} {
		iterator := RealAmplitudeIterator{
			Freq: NewFFTW(test.DomainSize).Freq,
			End:  pfutil.ProdInt(test.DomainSize),
		}

		num := 0
		for j := iterator.Next(); j != -1; j = iterator.Next() {
			num++
		}

		if num != test.ExpectNum {
			t.Errorf("Test #%d: Expected %d items got %d\n", i, test.ExpectNum, num)
		}
	}
}

func TestSortFactors(t *testing.T) {
	for i, test := range []struct {
		expr   string
		expect string
	}{
		{
			expr:   "solute*conc*temperature",
			expect: "conc*solute*temperature",
		},
		{
			expr:   "solute",
			expect: "solute",
		},
		{
			expr:   "current^2*voltage",
			expect: "current^2*voltage",
		},
		{
			expr:   "voltage*current^2",
			expect: "current^2*voltage",
		},
	} {
		res := SortFactors(test.expr)
		if res != test.expect {
			t.Errorf("Test #%d: Expected %s got %s\n", i, test.expect, res)
		}
	}
}
