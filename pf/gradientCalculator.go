package pf

import (
	"fmt"
	"math"

	"github.com/davidkleiven/gopf/pfutil"
)

// GradientCalculator calculates the gradient of a field
type GradientCalculator struct {
	FT          FourierTransform
	Comp        int
	KeepNyquist bool
}

// Calculate calculates the gradient of the data passed
// data contain the field in real-space
func (g *GradientCalculator) Calculate(indata []complex128, data []complex128) {
	copy(data, indata)
	g.FT.FFT(data)
	for i := range data {
		f := g.FT.Freq(i)[g.Comp]
		if math.Abs(f-0.5) < 1e-10 && !g.KeepNyquist {
			f = 0.0
		}
		data[i] *= complex(0.0, 2.0*math.Pi*f)
	}
	g.FT.IFFT(data)
	pfutil.DivRealScalar(data, float64(len(data)))
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

// DivGrad is a type used to represent the term Div F(c)Grad <field>
type DivGrad struct {
	Field string
	F     GenericFunction
}

// FuncName returns the name of the generic function F
func (dg *DivGrad) FuncName() string {
	return fmt.Sprintf("DivGrad_%s_Func", dg.Field)
}

// GradName returns the name of the gradient
func (dg *DivGrad) GradName(comp int) string {
	return fmt.Sprintf("GRAD_%s_%d", dg.Field, comp)
}

// PrepareModel impose nessecary changes in the model in order to use the
// DivGrad term. N is the number of grid points in the simulation domain,
// dim is the dimension of the simulation domain. FT is a fourier transformer
// required for gradient evaluations
func (dg *DivGrad) PrepareModel(N int, m *Model, FT FourierTransform) {
	dim := len(FT.Freq(0))
	for d := 0; d < dim; d++ {
		grad := GradientCalculator{
			FT:          FT,
			Comp:        d,
			KeepNyquist: false,
		}
		dField := grad.ToDerivedField(dg.GradName(d), N, m.Bricks[dg.Field])
		m.RegisterDerivedField(dField)

		dField2 := DerivedField{
			Name: dg.FuncName() + dg.GradName(d),
			Data: make([]complex128, N),
			Calc: func(data []complex128) {
				for i := range data {
					data[i] = dg.F(i, m.Bricks) * m.Bricks[dg.GradName(grad.Comp)].Get(i)
				}
			},
		}
		m.RegisterDerivedField(dField2)
	}
}

// Construct builds the right hand side term
func (dg *DivGrad) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		pfutil.Clear(field)
		dim := len(freq(0))
		for d := 0; d < dim; d++ {
			brick := bricks[dg.FuncName()+dg.GradName(d)]
			for i := range field {
				f := freq(i)[d]
				field[i] += complex(0.0, 2.0*math.Pi*f) * brick.Get(i)
			}
		}
	}
}

// OnStepFinished does nothing as we don't need any updates in between steps
func (dg *DivGrad) OnStepFinished(t float64) {}

// WeightedLaplacian is a type used to represent terms of the form
// F(c) LAP <field>, where F is a function of all the fields.
type WeightedLaplacian struct {
	// Name of the field to the right of the laplacian (must be registered in the model)
	Field string

	// Name of the function in front of the laplacian (must be registered in the model)
	PreFactor string

	// FourierTransformer of the correct domain size. This is used to
	// go back and fourth between the real domain and the fourier domain
	// internally
	FT FourierTransform
}

// Construct returns the fourier transformed representation of this term
func (wl *WeightedLaplacian) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		fieldBrick := bricks[wl.Field]

		// Calcualte real space laplacian of the field
		lap := LaplacianN{
			Power: 1,
		}

		for i := range field {
			field[i] = fieldBrick.Get(i)
		}
		lap.Eval(freq, field)

		wl.FT.IFFT(field)
		pfutil.DivRealScalar(field, float64(len(field)))

		// Calculate the realspace representation of the prefactor
		work := make([]complex128, len(field))
		funcBrick := bricks[wl.PreFactor]

		for i := range work {
			work[i] = funcBrick.Get(i)
		}
		wl.FT.IFFT(work)
		pfutil.DivRealScalar(work, float64(len(work)))

		for i := range field {
			field[i] *= work[i]
		}

		wl.FT.FFT(field)
	}
}
