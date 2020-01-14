package pf

import (
	"github.com/davidkleiven/gopf/elasticity"
	"github.com/davidkleiven/gosfft/sfft"
	"gonum.org/v1/gonum/mat"
)

// Indicator is the indicator functions used to distinguish the two phases
func Indicator(x float64) float64 {
	return 3.0*x*x - 2.0*x*x*x
}

// IndicatorDeriv is the derivative of the indicator function
func IndicatorDeriv(x float64) float64 {
	return 6.0*x - 6.0*x*x
}

// DisplacementGetter is a functon that type that returns the displacement, when given fourier
// transformed body forces (force), a corresponding frequency getter and a set elasticity tensor
type DisplacementGetter func(force [][]complex128, freq elasticity.Frequency, matProp elasticity.Rank4Tensor) [][]complex128

// HomogeneousModulusLinElast is a type that is used in phase field models where the elastic
// constants are homogeneous throughout the domain. In other words it represents the energy
// term
// E = (1/2)*C_{jikl}(eps_{ij} - eps^*_{ij}H(x))(eps_{kl} - eps^*_{kl}H(x))
// where C is the elastic tensor, eps^* is the misfit strain and x is a field that is one if
// we are inside the domain where misfit strains exists and zero elswhere. The specific name
// of the name of the field x is stored in FieldName. When this term is used in an equation it
// is assumed that the right hand side consists of dx/dt = -dE/dx + ..., i.e. the time evolution
// of the field should be such that it minimizes the strain energy. The strain eps_{ij} is
// determined enforcing mechanical equillibrium
type HomogeneousModulusLinElast struct {
	FieldName string
	Field     []float64
	Misfit    *mat.Dense
	EffForce  elasticity.EffectiveForce
	MatProp   elasticity.Rank4
	Disps     DisplacementGetter
	FT        FourierTransform
	Dim       int
	N         int
}

// Construct returns the function needed to build the term on the
// right hand side
func (h *HomogeneousModulusLinElast) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) []complex128 {
		for i := range field {
			field[i] = complex(0.0, 0.0)
		}

		work := make([]complex128, h.N)
		for i := range work {
			work[i] = complex(Indicator(h.Field[i]), 0.0)
		}
		h.FT.FFT(work)

		force := h.Force(work)
		disp := h.Disps(force, h.Freq, &h.MatProp)

		// Fill work with the derivative of the indicator
		for i := range work {
			work[i] = complex(IndicatorDeriv(h.Field[i]), 0.0)
		}

		A := h.MatProp.ContractLast(h.Misfit)

		for i := 0; i < h.Dim; i++ {
			for j := i; j < h.Dim; j++ {
				strains := elasticity.Strain(disp, h.Freq, i, j)
				h.FT.IFFT(strains) // Obtain real-space strains
				DivRealScalar(strains, float64(len(strains)))
				ElemwiseMul(strains, work)
				h.FT.FFT(strains)

				factor := 1.0
				if i != j {
					factor *= 2.0
				}

				for k := range field {
					field[k] += complex(factor*A.At(i, j), 0.0) * strains[k]
				}
			}
		}

		eDensity := elasticity.EnergyDensity(&h.MatProp, h.Misfit)

		// Fill work with indicator times indicatorDerov
		for i := range work {
			work[i] = complex(Indicator(h.Field[i])*IndicatorDeriv(h.Field[i]), 0.0)
		}
		h.FT.FFT(work)
		for k := range field {
			field[k] -= complex(2.0*eDensity, 0.0) * work[k]
		}
		return field
	}
}

// Freq wraps the passed frequency method such that the length of the returned frequency
// is always 3
func (h *HomogeneousModulusLinElast) Freq(i int) []float64 {
	if h.Dim == 3 {
		return h.FT.Freq(i)
	}
	res := make([]float64, 3)
	f := h.FT.Freq(i)
	copy(res[:h.Dim], f)
	return res
}

// Force returns the effective force
func (h *HomogeneousModulusLinElast) Force(indicator []complex128) [][]complex128 {
	res := make([][]complex128, h.N)
	for i := range res {
		res[i] = make([]complex128, 3)
	}

	for i := 0; i < h.Dim; i++ {
		force := h.EffForce.Get(i, h.Freq, indicator)
		for j := range force {
			res[j][i] = force[j]
		}
	}
	return res
}

// OnStepFinished update the real space version of the field
func (h *HomogeneousModulusLinElast) OnStepFinished(t float64, bricks map[string]Brick) {
	for i := range h.Field {
		h.Field[i] = real(bricks[h.FieldName].Get(i))
	}
}

// NewHomogeneousModolus initializes a new instance of the linear elasticity model
func NewHomogeneousModolus(fieldName string, domainSize []int, matProp elasticity.Rank4, misfit *mat.Dense) *HomogeneousModulusLinElast {
	linElast := HomogeneousModulusLinElast{
		FieldName: fieldName,
		Dim:       len(domainSize),
		N:         ProdInt(domainSize),
		MatProp:   matProp,
		Misfit:    misfit,
		EffForce:  elasticity.NewEffectiveForceFromMisfit(matProp, misfit),
		Field:     make([]float64, ProdInt(domainSize)),
		Disps:     elasticity.Displacements,
	}
	dim := len(domainSize)
	if dim == 2 {
		linElast.FT = sfft.NewFFT2(domainSize[0], domainSize[1])
	} else if dim == 3 {
		linElast.FT = sfft.NewFFT3(domainSize[0], domainSize[1], domainSize[2])
	} else {
		panic("pf: Domain size has to be either of length 2 or length 3")
	}
	return &linElast
}
