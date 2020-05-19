package pf

import (
	"math"
	"math/cmplx"

	"github.com/davidkleiven/gopf/pfutil"
	"github.com/davidkleiven/gosfft/sfft"
)

// SquaredGradient an be used for terms that has a gradient raised to second
// power (e.g (grad u)^2, where u is a field). The power has to be an even number
// Factor is multiplied with the gradient. Thus, the full term reads
// Factor*(grad Field)^2
type SquaredGradient struct {
	Field  string
	FT     FourierTransform
	Factor float64
}

// NewSquareGradient returns a new instance of square gradient
func NewSquareGradient(field string, domainSize []int) SquaredGradient {
	sq := SquaredGradient{
		Field:  field,
		Factor: 1.0,
	}
	if len(domainSize) == 2 {
		sq.FT = sfft.NewFFT2(domainSize[0], domainSize[1])
	} else if len(domainSize) == 3 {
		sq.FT = sfft.NewFFT3(domainSize[0], domainSize[1], domainSize[2])
	} else {
		panic("squaregradient: Domain size has to be of length 2 or 3")
	}
	return sq
}

// Construct return a function the calculates the squared of the gradient of
func (s *SquaredGradient) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		pfutil.Clear(field)
		work := make([]complex128, len(field))
		k := freq(0)
		dim := len(k)
		tol := 1e-10
		for d := 0; d < dim; d++ {
			for i := range work {
				f := freq(i)
				if math.Abs(f[d]-0.5) < tol {
					f[d] = 0.0
				}
				work[i] = bricks[s.Field].Get(i) * complex(0.0, 2.0*math.Pi*f[d])
			}

			s.FT.IFFT(work)
			for i := range work {
				work[i] = cmplx.Pow(work[i]/complex(float64(len(work)), 0.0), 2.0)
			}
			s.FT.FFT(work)

			for i := range work {
				field[i] += work[i] * complex(s.Factor, 0.0)
			}
		}
	}
}

// OnStepFinished does not need to perform any work
func (s *SquaredGradient) OnStepFinished(t float64, bricks map[string]Brick) {}
