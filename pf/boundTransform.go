package pf

import "math"

// BoundTransform is a type that maps variables on the real-line (-inf, inf) into the
// interval (Min, Max). It does so by applying the following transformation:
// y = Min + (Max-Min)*(arctan(x)/pi + 1)
// x is a number on the real line.
type BoundTransform struct {
	Min float64
	Max float64
}

// Forward returns the mapped coordinate
func (bt *BoundTransform) Forward(x float64) float64 {
	return bt.Min + (bt.Max-bt.Min)*(math.Atan(x)/math.Pi+0.5)
}

// Backward maps the coordinate y into x
func (bt *BoundTransform) Backward(y float64) float64 {
	return math.Tan(math.Pi*(y-bt.Min)/(bt.Max-bt.Min) - 0.5*math.Pi)
}
