package pf

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"regexp"
	"strconv"
	"strings"
)

// Term is generic function type that evaluates the right hand side of a set of
// ODE used to evolve the phase fields. The function should place the values into
// the field array
type Term func(freq Frequency, t float64, field []complex128)

// RHS is a struct used to represent the "right-hand-side" of a set of ODE
type RHS struct {
	Terms []Term
	Denum []Term
}

// Build constructs the right-hand-side of an equation based on a string
// representation
func Build(eq string, m *Model) RHS {
	sides := strings.Split(eq, "=")
	if len(sides) != 2 {
		panic("build: equality sign can only occur once")
	}

	field := fieldNameFromLeibniz(sides[0])
	termsStr := SplitOnMany(sides[1], []string{"+", "-"})
	var rhs RHS
	for _, t := range termsStr {
		name := removeKnownPrefixes(t.SubString)
		prefixes := getKnownPrefixes(t.SubString)
		prefixes = append(prefixes, t.PreceedingDelimiter)
		if m.IsImplicitTerm(name) {
			rhs.Denum = append(rhs.Denum, constructFunc(m.ImplicitTerms[name].Construct(m.Bricks), prefixes))
		} else if m.IsExplicitTerm(name) {
			rhs.Terms = append(rhs.Terms, constructFunc(m.ExplicitTerms[name].Construct(m.Bricks), prefixes))
		} else if m.IsMixedTerm(name) {
			rhs.Denum = append(rhs.Denum, constructFunc(m.MixedTerms[name].ConstructLinear(m.Bricks), prefixes))
			rhs.Terms = append(rhs.Terms, constructFunc(m.MixedTerms[name].ConstructNonLinear(m.Bricks), prefixes))
		} else if isBilinear(t.SubString, field, m.AllFieldNames()) {
			t.SubString = strings.Replace(t.SubString, field, "", -1)
			rhs.Denum = append(rhs.Denum, ConcreteTerm(t, m))
		} else {
			rhs.Terms = append(rhs.Terms, ConcreteTerm(t, m))
		}
	}
	return rhs
}

// fieldNameFromLeibniz extracts a field name from a Leibniz formatted
// differnetation operation (e.g dkappa/dt)
func fieldNameFromLeibniz(leibniz string) string {
	if len(leibniz) <= 3 {
		panic("rhsbuilder: Length of the Leibniz formatted string has to be at least 3")
	}
	if leibniz[0:1] != "d" || leibniz[len(leibniz)-3:] != "/dt" {
		panic("rhsbuilder: Passed string is not a leibniz formatted string")
	}
	return leibniz[1 : len(leibniz)-3]
}

// isBilinear checks if the term given is bilinear in the passed field
func isBilinear(term string, field string, fieldNames []string) bool {
	fieldReg := regexp.MustCompile(field)
	resCount := fieldReg.FindAllStringIndex(term, -1)
	if len(resCount) != 1 {
		return false
	}

	// Check if any of ther other fields are present. In that case, the term
	// will not be linear in the fourier domain
	for _, f := range fieldNames {
		if f == field {
			continue
		}
		r := regexp.MustCompile(f)
		count := r.FindAllStringIndex(term, -1)
		if len(count) > 0 {
			return false
		}
	}

	// Match the field name until * or / is found
	regIncludingPowers := regexp.MustCompile(field + "*[^/\\*]*")
	res := regIncludingPowers.FindString(term)

	// Extract an power (if exists) (the number after ^)
	regPower := regexp.MustCompile("\\^(-?\\d+\\.?\\d*)")
	strPower := regPower.FindStringSubmatch(res)
	if len(strPower) <= 1 {
		return true
	}
	power, err := strconv.ParseFloat(strPower[1], 64)

	if err != nil || math.Abs(power-1.0) < 1e-10 {
		// No power or equal to 1
		return true
	}
	return false
}

// ValidName returns true if the parser knows how to parse it
func ValidName(name string, model *Model) bool {
	nameStripped := strings.ReplaceAll(name, " ", "")
	specialNames := []string{"", "LAP"}
	for _, n := range specialNames {
		if nameStripped == n {
			return true
		}
	}

	if nameStripped[:3] == "LAP" {
		nameStripped = nameStripped[3:]
	}
	return model.IsBrickName(nameStripped) || model.IsFieldName(nameStripped)
}

// ConcreteTerm returns a function representing the passed term
func ConcreteTerm(termDelim SubStringDelimiter, m *Model) Term {
	term := termDelim.SubString
	sign := 1.0
	if termDelim.PreceedingDelimiter == "-" {
		sign = -1.0
	}

	fieldReg := regexp.MustCompile("[^\\*]*")
	res := fieldReg.FindAllStringSubmatch(term, -1)

	brickNames := []string{}
	powers := []float64{}

	nameNoPow := regexp.MustCompile("^[^\\^]*")
	for i := range res {
		name := nameNoPow.FindString(res[i][0])

		if !m.IsFieldName(name) && m.IsBrickName(name) {
			brickNames = append(brickNames, name)
			powers = append(powers, GetPower(res[i][0]))
		} else if !ValidName(name, m) {
			msg := fmt.Sprintf("rhsBuilder: Name %s is not defined!", name)
			panic(msg)
		}
	}

	fieldName := GetFieldName(SortFactors(term), m.AllFieldNames())

	if strings.Contains(term, "LAP") {
		// Term with Laplace operator
		lapWithPowReg := regexp.MustCompile("LAP*[^a-zA-Z]*")
		res := lapWithPowReg.FindString(term)
		lap := LaplacianN{Power: int(GetPower(res))}
		return func(freq Frequency, t float64, field []complex128) {
			for i := range field {
				field[i] = complex(sign, 0.0)
				for j := range brickNames {
					field[i] *= cmplx.Pow(m.Bricks[brickNames[j]].Get(i), complex(powers[j], 0.0))
				}
			}

			if fieldName != "" {
				for i := range field {
					field[i] *= m.Bricks[fieldName].Get(i)
				}
			}
			lap.Eval(freq, field)
		}
	}

	// Term with out laplacian operators
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			field[i] = complex(sign, 0.0)
			for j := range brickNames {
				field[i] *= cmplx.Pow(m.Bricks[brickNames[j]].Get(i), complex(powers[j], 0.0))
			}
		}

		if fieldName != "" {
			for i := range field {
				field[i] *= m.Bricks[fieldName].Get(i)
			}
		}
	}
}

// flipSign multiplies all items in the passed array by -1
func flipSign(data []complex128) {
	for i := range data {
		data[i] *= -1.0
	}
}

func constructFunc(term Term, prefixes []string) Term {
	var f Term
	if len(prefixes) == 0 {
		return term
	}

	switch prefixes[0] {
	case "-":
		f = func(freq Frequency, t float64, field []complex128) {
			term(freq, t, field)
			flipSign(field)
		}
		break
	case "LAP^4":
		f = func(freq Frequency, t float64, field []complex128) {
			lap := LaplacianN{Power: 4}
			term(freq, t, field)
			lap.Eval(freq, field)
		}
		break
	case "LAP^2":
		f = func(freq Frequency, t float64, field []complex128) {
			lap := LaplacianN{Power: 2}
			term(freq, t, field)
			lap.Eval(freq, field)
		}
		break
	case "LAP":
		f = func(freq Frequency, t float64, field []complex128) {
			lap := LaplacianN{Power: 1}
			term(freq, t, field)
			lap.Eval(freq, field)
		}
		break
	case " ", "+":
		// Do nothing and do not print message that the prefix is unrecognezed
		f = term
		break
	default:
		log.Printf("Unrecognized prefix %s (potential problem)", prefixes[0])
		f = term
	}
	return constructFunc(f, prefixes[1:])
}

func knownPrefixes() []string {
	return []string{
		"-",
		"LAP^4",
		"LAP^2",
		"LAP",
		"*",
	}
}

func removeKnownPrefixes(str string) string {
	prefixes := knownPrefixes()
	return recursiveRemove(prefixes, str)
}

func getKnownPrefixes(str string) []string {
	pref := []string{}
	prefixes := knownPrefixes()
	for len(prefixes) > 0 {
		prefix := prefixes[0]
		prefixes = prefixes[1:]
		if strings.HasPrefix(str, prefix) {
			pref = append(pref, prefix)
			prefixes = knownPrefixes()
			str = str[len(prefix):]
		}
	}
	return pref
}

func recursiveRemove(prefixes []string, str string) string {
	if len(prefixes) == 0 {
		return str
	}
	prefix := prefixes[0]
	prefixes = prefixes[1:]
	if strings.HasPrefix(str, prefix) {
		str = str[len(prefix):]

		// Reset prefix list in case there are more of the same
		prefixes = knownPrefixes()
	}
	return recursiveRemove(prefixes, str)
}
