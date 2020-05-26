package pf

import (
	"fmt"
	"math"
	"math/cmplx"
	"regexp"
	"strconv"
	"strings"

	"github.com/davidkleiven/gopf/pfutil"
)

// GetNonLinearFieldExpressions returns constructions that must be FFT separateley. Note
// that if the expression is bilinear in field, field is omitted from the returned expression
func GetNonLinearFieldExpressions(pattern string, field string, fieldNames []string) string {
	expr := ""
	for i := range fieldNames {
		if fieldNames[i] == field && isBilinear(pattern, field, fieldNames) {
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

// ModalFilter is a generic interface for modal filters
type ModalFilter interface {
	Eval(x float64) float64
}

// ApplyModalFilter applies the filter f in-place to data
func ApplyModalFilter(filter ModalFilter, freq Frequency, data []complex128) {
	for i := range data {
		f := freq(i)
		fRad := math.Sqrt(pfutil.Dot(f, f))
		value := fRad * 2.0 / math.Pi
		data[i] *= complex(filter.Eval(value), 0.0)
	}
}

// SubStringDelimiter is a type that represents a substring as well as
// the delimiter preceeding it in the string it was extracted from
type SubStringDelimiter struct {
	SubString           string
	PreceedingDelimiter string
}

// SplitOnMany splits a string on several delimiters. The resulted split is
// returned together with the delimiter preceeding it
func SplitOnMany(value string, delimiters []string) []SubStringDelimiter {
	substrings := []SubStringDelimiter{}

	queue := []SubStringDelimiter{{
		SubString:           value,
		PreceedingDelimiter: "",
	}}

	allDelims := ""
	for _, delim := range delimiters {
		allDelims += delim
	}

	for len(queue) > 0 {
		currentSubString := queue[0]
		queue = queue[1:]

		if !strings.ContainsAny(currentSubString.SubString, allDelims) {
			substrings = append(substrings, currentSubString)
			continue
		}

		delim := delimiters[0]
		for i := range delimiters {
			if strings.Contains(currentSubString.SubString, delimiters[i]) {
				delim = delimiters[i]
				break
			}
		}

		splits := strings.Split(currentSubString.SubString, delim)
		queue = append(queue, SubStringDelimiter{
			SubString:           splits[0],
			PreceedingDelimiter: currentSubString.PreceedingDelimiter,
		})

		for i := 1; i < len(splits); i++ {
			queue = append(queue, SubStringDelimiter{
				SubString:           splits[i],
				PreceedingDelimiter: delim,
			})
		}
	}
	return substrings
}

// UniqueFreqIterator is an iterator that can be used to iterate over all the unique
// frequencies of the fourier transform of a real-valued function. This iterator can
// be used in a for loop as follows
// for i := iterator.Next(); i != -1; i = iterator.Next()
type UniqueFreqIterator struct {
	Freq Frequency
	End  int
	next int
}

// Next returns the index of unique site. It returns -1 when exhausted
func (ufi *UniqueFreqIterator) Next() int {
	for i := ufi.next; i < ufi.End; i++ {
		f := ufi.Freq(i)

		// Special case last frequency is 0 or 0.5
		//if f[len(f)-1] > -tol {
		if includeFreq(f, len(f)-1) {
			ufi.next = i + 1
			return i
		}
	}
	return -1
}

func includeFreq(f []float64, split int) bool {
	tol := 1e-10
	if split == 0 {
		return f[0] > -tol
	}

	if math.Abs(f[split]) < tol || math.Abs(f[split]) > 0.5-tol {
		return includeFreq(f, split-1)
	}
	return f[split] > -tol
}

// firstFreqsAreNyquistOrZero returns true if all frequencies (apart from the last)
// is either 0 or 0.5
func (ufi *UniqueFreqIterator) firstFreqsAreNyquistOrZero(f []float64) bool {
	tol := 1e-10
	for j := range f {
		if math.Abs(f[j]) > tol || math.Abs(f[j]) < 0.5-tol {
			return false
		}
	}
	return true
}

// RealAmplitudeIterator iterates over all frequencies that has a real fourier amplitude
// when the input signal has a real amplitude
type RealAmplitudeIterator struct {
	Freq Frequency
	End  int
	next int
}

// Next returns the next index. It returns -1 when the iterator is exhausted
func (rai *RealAmplitudeIterator) Next() int {
	tol := 1e-10
	for i := rai.next; i < rai.End; i++ {
		f := rai.Freq(i)

		allZeroOrNyquist := true
		for j := range f {
			if math.Abs(f[j]) > tol && math.Abs(f[j]) < 0.5-tol {
				allZeroOrNyquist = false
			}
		}

		if allZeroOrNyquist {
			fmt.Printf("%v\n", f)
			rai.next = i + 1
			return i
		}
	}
	return -1
}
