package pfc

import (
	"math"
	"github.com/davidkleiven/gopf/pfutil"
)

// 3*x^2 - 2x^3
func Indicator(x float64) float64 {
	if x < 0.0 {
		return 0.0
	} else if x > 1{
		return 1.0
	} 
	return 3 * x * x - 2 * x * x * x
}

func IndicatorDeriv(x float64) float64 {
	if x < 0.0 || x > 1.0 {
		return 0.0
	}
	return 6 * x - 6 * x * x
}

type TwoPhaseCorrFunc struct {
	CorrFunc []ReciprocalSpacePairCorrelation
}

func Dot(x []float64, y []float64) float64 {
	res := 0.0
	for i := range x {
		res += x[i]*y[i]
	}
	return res
}

func (tpcf *TwoPhaseCorrFunc) Evaluate(cFT []complex128, fft *pfutil.FFTWWrapper) []complex128 {
	if len(tpcf.CorrFunc) != 2 {
		panic("Length of CorrFunc must be 2\n")
	}

	// Calcualte real space correlation functions
	corrFunc1FT := make([]complex128, len(cFT))
	corrFunc2FT := make([]complex128, len(cFT))
	twoPi := 2.0*math.Pi
	for i := range corrFunc1FT {
		freq := fft.Freq(i)
		k := twoPi*math.Sqrt(Dot(freq, freq))
		corrFunc1FT[i] = complex(tpcf.CorrFunc[0].Eval(k), 0.0)
		corrFunc2FT[i] = complex(tpcf.CorrFunc[1].Eval(k), 0.0)
	}

	corrFunc1 := fft.IFFT(corrFunc1FT)
	corrFunc2 := fft.IFFT(corrFunc2FT)
	pfutil.DivRealScalar(corrFunc1, float64(len(corrFunc1)))
	pfutil.DivRealScalar(corrFunc2, float64(len(corrFunc2)))

	// Calculate real space concentration
	conc := fft.IFFT(cFT)
	pfutil.DivRealScalar(conc, float64(len(conc)))
	
	// C_eff = Indicator(conc)*CorrFunc1 + (1-Indicator(conc))*CorrFunc2
	C_eff := corrFunc1FT // Overwrite corrFunc1FT
	for i := range C_eff {
		x := real(conc[i])
		C_eff[i] = complex(Indicator(x), 0.0) * corrFunc1[i] + complex(1-Indicator(x), 0.0)*corrFunc2[i]
	}
	return C_eff
}

func (tpcf *TwoPhaseCorrFunc) EvaluateFT(cFT []complex128, fft *pfutil.FFTWWrapper) []complex128 {
	Ceff := tpcf.Evaluate(cFT, fft)
	fft.FFT(Ceff)
	return Ceff
}