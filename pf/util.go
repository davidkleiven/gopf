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

// ElemwiseMul multiplies dst and data and places the result in dst.
func ElemwiseMul(dst []complex128, data []complex128) {
	for i := range dst {
		dst[i] *= data[i]
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

// RealPartAsUint8 return the real part of the field as uint8. The data is scaled
// such that min --> 0 and max --> 255
func RealPartAsUint8(data []complex128, min float64, max float64) []uint8 {
	res := make([]uint8, len(data))
	if math.Abs(max-min) < 1e-10 {
		max = min + 1.0
	}
	for i := range data {
		res[i] = uint8(255 * (real(data[i]) - min) / (max - min))
	}
	return res
}

// MaxReal calculates the maximum real part
func MaxReal(data []complex128) float64 {
	maxval := real(data[0])
	for i := range data {
		if real(data[i]) > maxval {
			maxval = real(data[i])
		}
	}
	return maxval
}

// MinReal returns the minimum real-part value
func MinReal(data []complex128) float64 {
	minval := real(data[0])
	for i := range data {
		if real(data[i]) < minval {
			minval = real(data[i])
		}
	}
	return minval
}

// nodeIdx2 returns the node index for 2D grid
func nodeIdx2(domainSize []int, idx []int) int {
	return idx[0]*domainSize[1] + idx[1]
}

// nodeIdx3 returns the node index for 3D grid
func nodeIdx3(domainSize []int, idx []int) int {
	return idx[2]*domainSize[0]*domainSize[1] + idx[0]*domainSize[1] + idx[1]
}

// NodeIdx returns the index of the node corresponding to a given typle of index
func NodeIdx(domainSize []int, idx []int) int {
	if len(domainSize) == 2 && len(idx) == 2 {
		return nodeIdx2(domainSize, idx)
	} else if len(domainSize) == 3 && len(idx) == 3 {
		return nodeIdx3(domainSize, idx)
	}
	panic("util: Domain size and idx has to be of length 2 or 3")
}

func pos3(domainSize []int, nodeNum int) []int {
	col := nodeNum % domainSize[1]
	row := (nodeNum / domainSize[1]) % domainSize[0]
	depth := nodeNum / (domainSize[0] * domainSize[1])
	return []int{row, col, depth}
}

func pos2(domainSize []int, nodeNum int) []int {
	col := nodeNum % domainSize[1]
	row := nodeNum / domainSize[1]
	return []int{row, col}
}

// Pos converts the node number to position
func Pos(domainSize []int, nodeNum int) []int {
	if len(domainSize) == 2 {
		return pos2(domainSize, nodeNum)
	} else if len(domainSize) == 3 {
		return pos3(domainSize, nodeNum)
	}
	panic("util: Domain size has to be either 2 or 3")
}

// Clear sets all elements in the slice to zero
func Clear(data []complex128) {
	for i := range data {
		data[i] = complex(0.0, 0.0)
	}
}

// ModalFilter is a generic interface for modal filters
type ModalFilter interface {
	Eval(x float64) float64
}

// ApplyModalFilter applies the filter f in-place to data
func ApplyModalFilter(filter ModalFilter, freq Frequency, data []complex128) {
	for i := range data {
		f := freq(i)
		fRad := math.Sqrt(Dot(f, f))
		value := fRad * 2.0 / math.Pi
		data[i] *= complex(filter.Eval(value), 0.0)
	}
}
