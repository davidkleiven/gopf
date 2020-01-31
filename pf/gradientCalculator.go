package pf

import "math"

// GradientCalculator calculates the gradient of a field
type GradientCalculator struct {
	FT   FourierTransform
	Comp int
}

// Calculate calculates the gradient of the data passed
// data contain the field in real-space
func (g *GradientCalculator) Calculate(indata []complex128, data []complex128) {
	copy(data, indata)
	g.FT.FFT(data)
	for i := range data {
		f := g.FT.Freq(i)[g.Comp]
		if math.Abs(f-0.5) < 1e-10 {
			f = 0.0
		}
		data[i] *= complex(0.0, 2.0*math.Pi*f)
	}
	g.FT.IFFT(data)
	DivRealScalar(data, float64(len(data)))
}

// ToDerivedField constructs a derived field from the gradient calculator.
// N is the number of grid points and brick is the brick that should be
// differentiated
func (g *GradientCalculator) ToDerivedField(name string, N int, brick Brick) DerivedField {
	return DerivedField{
		Data: make([]complex128, N),
		Name: name,
		Calc: func(data []complex128) {
			for i := range data {
				data[i] = brick.Get(i)
			}
			g.Calculate(data, data)
		},
	}
}
