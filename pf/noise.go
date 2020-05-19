package pf

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/davidkleiven/gopf/pfutil"
)

// WhiteNoise is a type that can be used to add white noise to a model
// if Y(x, t) is a noise term, the correlation function is defined by
// <Y(x, t)Y(x', t')> = 2*Strength*delta(x-x')delta(t-t') (e.g. uncorrelated
// in time and space)
type WhiteNoise struct {
	Strength float64
}

// Generate returns random gaussian distributed noise
func (w *WhiteNoise) Generate(i int, bricks map[string]Brick) complex128 {
	std := math.Sqrt(2.0 * w.Strength)
	return complex(rand.NormFloat64()*std, 0.0)
}

// ConservativeNoise adds noise in a such a way that the field it is added to is conserved
type ConservativeNoise struct {
	UniquePrefix uint32
	Strength     float64
	Dim          int
}

// NewConservativeNoise returns an instance of ConservativeNoise with a
// correctly initialized UniquePrefix which is used to identify derived fields
// associated with the conservative noise
func NewConservativeNoise(strength float64, dim int) ConservativeNoise {
	return ConservativeNoise{
		UniquePrefix: rand.Uint32(),
		Strength:     strength,
		Dim:          dim,
	}
}

// GetCurrentName returns the name of the current field
func (cn *ConservativeNoise) GetCurrentName(comp int) string {
	return fmt.Sprintf("%d_current_%d", cn.UniquePrefix, comp)
}

// CurrentFieldsAreRegistered checks that all the current fields are
// available among the bricks
func (cn *ConservativeNoise) CurrentFieldsAreRegistered(bricks map[string]Brick) bool {
	for comp := 0; comp < cn.Dim; comp++ {
		if _, ok := bricks[cn.GetCurrentName(comp)]; !ok {
			return false
		}
	}
	return true
}

// Construct builds the right hand side term
func (cn *ConservativeNoise) Construct(bricks map[string]Brick) Term {
	if !cn.CurrentFieldsAreRegistered(bricks) {
		panic("ConservativeCurrent: Current fields are not register. Make sure that you have registered the fields returned by RequiredDerivedFields.")
	}
	return func(freq Frequency, t float64, field []complex128) {
		pfutil.Clear(field)
		for comp := 0; comp < cn.Dim; comp++ {
			brick := bricks[cn.GetCurrentName(comp)]
			for i := range field {
				f := freq(i)[comp]
				if math.Abs(math.Abs(f)-0.5) > 1e-6 {
					omegaHalf := math.Pi * f
					field[i] += complex(0.0, 2.0*math.Sin(omegaHalf)) * brick.Get(i)
				}
			}
		}
	}
}

// OnStepFinished does nothing
func (cn *ConservativeNoise) OnStepFinished(t float64, bricks map[string]Brick) {}

// RequiredDerivedFields returns the a set of derived fields that is nessecary to register
// in order to use conservative noise. The number of nodes in the grid is specified via
// the numNodes argument
func (cn *ConservativeNoise) RequiredDerivedFields(numNodes int) []DerivedField {
	dfields := make([]DerivedField, cn.Dim)
	for i := 0; i < cn.Dim; i++ {
		dfields[i] = DerivedField{
			Data: make([]complex128, numNodes),
			Name: cn.GetCurrentName(i),
			Calc: func(data []complex128) {
				std := math.Sqrt(2.0 * cn.Strength)
				for i := range data {
					data[i] = complex(rand.NormFloat64()*std, 0.0)
				}
			},
		}
	}
	return dfields
}
