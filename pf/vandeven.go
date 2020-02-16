package pf

import "math"

// Vandeven implements the family of filters described in
// Vandeven, H., 1991. Family of spectral filters for discontinuous problems.
// Journal of Scientific Computing, 6(2), pp.159-192.
type Vandeven struct {
	Data []float64
}

// NewVandeven constructs a new vanden filter of the passed order
func NewVandeven(order int) Vandeven {
	var filter Vandeven
	filter.Data = make([]float64, 1000)
	filter.Data[0] = 1.0
	prefactor := math.Gamma(float64(2*order)) / (math.Gamma(float64(order)) * math.Gamma(float64(order)))
	dx := 1.0 / 999.0
	for i := 1; i < 1000; i++ {
		x := float64(i) * dx
		i2 := math.Pow(x*(1-x), float64(order-1))
		x1 := x - dx
		i1 := math.Pow(x1*(1.0-x1), float64(order-1))
		filter.Data[i] = filter.Data[i-1] - prefactor*0.5*(i1+i2)*dx
	}
	return filter
}

// Eval evaluates the filter at x
func (v *Vandeven) Eval(x float64) float64 {
	N := float64(len(v.Data) - 1)
	idx := int(x * N)
	if idx >= len(v.Data)-1 {
		return v.Data[len(v.Data)-1]
	}
	dx := 1.0 / N
	dy := v.Data[idx+1] - v.Data[idx]
	x0 := float64(idx) * dx
	return v.Data[idx] + (x-x0)*dy/dx
}
