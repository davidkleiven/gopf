package pf

import (
	"math"
	"math/rand"
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
