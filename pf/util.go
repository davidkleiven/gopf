package pf

import (
	"math"
	"math/cmplx"
	"regexp"
	"strconv"
	"strings"
)

// CmplxEqualApprox returns true if to complex arrays are equal within the passed tolerance
func CmplxEqualApprox(a []complex128, b []complex128, tol float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if math.Abs(real(a[i])-real(b[i])) > tol || math.Abs(imag(a[i])-imag(b[i])) > tol {
			return false
		}
	}
	return true
}

// GetNonLinearFieldExpressions returns constructions that must be FFT separateley. Note
// that if the expression is bilinear in field, field is omitted from the returned expression
func GetNonLinearFieldExpressions(pattern string, field string, fieldNames []string) string {
	expr := ""
	for i := range fieldNames {
		if fieldNames[i] == field && isBilinear(pattern, field) {
			continue
		}
		regField := regexp.MustCompile(fieldNames[i] + "[^\\*]*")
		res := regField.FindString(pattern)
		if res != "" {
			expr += res + "*"
		}
	}

	if len(expr) > 1 {
		return expr[:len(expr)-1]
	}
	return expr
}

// DerivedFieldCalcFromDesc returns a derived field calculator based on its description
// (e.g. conc^2*eta, if conc and eta are two fields)
func DerivedFieldCalcFromDesc(desc string, fields []Field) DerivedFieldCalc {
	fieldMap := make(map[string]*Field)

	for i := range fields {
		fieldMap[fields[i].Name] = &fields[i]
	}

	fieldReg := regexp.MustCompile("[^\\*]*")
	res := fieldReg.FindAllStringSubmatch(desc, -1)

	fieldNames := make([]string, len(res))
	powers := make([]float64, len(fieldNames))

	nameNoPow := regexp.MustCompile("^[^\\^]*")
	for i := range res {
		fieldNames[i] = nameNoPow.FindString(res[i][0])
		powers[i] = GetPower(res[i][0])
	}

	return func(data []complex128) {
		for i := range data {
			data[i] = 1.0
			for j := range fieldNames {
				data[i] *= cmplx.Pow(fieldMap[fieldNames[j]].Data[i], complex(powers[j], 0.0))
			}
		}
	}
}

// GetPower returns the power from a string
func GetPower(pattern string) float64 {
	regPower := regexp.MustCompile("\\^(-?\\d+\\.?\\d*)")
	strPower := regPower.FindStringSubmatch(pattern)

	if len(strPower) <= 1 {
		return 1.0
	}
	power, err := strconv.ParseFloat(strPower[1], 64)
	if err != nil {
		panic(err)
	}
	return power
}

// GetFieldName returns the field name of a term.
func GetFieldName(term string, fieldNames []string) string {
	field := ""
	for _, f := range fieldNames {
		if strings.Contains(term, f) {
			// Check that there are no other matches
			withoutField := strings.Replace(term, f, "", -1)

			ok := true
			for _, f1 := range fieldNames {
				if strings.Contains(withoutField, f1) {
					ok = false
					break
				}
			}

			if ok && len(f) > len(field) {
				field = f
			}
		}
	}
	return field
}

// ElemwiseAdd adds dst and data and places the result in dst
func ElemwiseAdd(dst []complex128, data []complex128) {
	for i := range dst {
		dst[i] += data[i]
	}
}

// DivRealScalar divides each element in the comlex array by a real scalar
func DivRealScalar(data []complex128, factor float64) []complex128 {
	cfactor := complex(factor, 0.0)
	for i := range data {
		data[i] /= cfactor
	}
	return data
}

// ProdInt calculates the product of all the elements in the passed sequence
func ProdInt(a []int) int {
	res := 1
	for i := range a {
		res *= a[i]
	}
	return res
}

// Dot calculates the dot product between two slices
func Dot(a []float64, b []float64) float64 {
	res := 0.0
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}
