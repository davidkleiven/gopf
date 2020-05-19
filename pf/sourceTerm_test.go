package pf

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
)

func demo(t float64) float64 {
	return 2.0 * t
}

func freq(i int) []float64 {
	return []float64{float64(i)}
}

func TestSourceTerm(t *testing.T) {
	src := NewSource([]float64{2.5}, demo)
	cData := make([]complex128, 2)
	src.Eval(freq, 2.0, cData)
	a1 := cmplx.Exp(-complex(0.0, 2.0*math.Pi*2.5*0.0))
	a2 := cmplx.Exp(-complex(0.0, 2.0*math.Pi*2.5*1.0))
	expect := []complex128{a1 * complex(4.0, 0.0), a2 * complex(4.0, 0.0)}

	if !pfutil.CmplxEqualApprox(expect, cData, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, cData)
	}
}
