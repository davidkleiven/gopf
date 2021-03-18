package pfc

import (
	"testing"
	"math"
	"github.com/davidkleiven/gopf/pfutil"
)

func TestIndicator(t *testing.T) {
	for i, test := range []struct {
		x float64
		want float64
	}{
		{
			x: -0.3,
			want: 0.0,
		},
		{
			x: 1.1,
			want: 1.0,
		},
		{
			x: 0.5,
			want: 0.5,
		},
	}{
		tol := 1e-6
		got := Indicator(test.x)
		if math.Abs(got - test.want) > tol {
			t.Errorf("Test #%d: Wanted %f got %f\n", i, test.want, got)
		}
	}
}

func TestIndicatorDeriv(t *testing.T) {
	for i, test := range []struct {
		x float64
		want float64
	}{
		{
			x: -0.4,
			want: 0.0, 
		},
	}{
		tol := 1e-6
		got := IndicatorDeriv(test.x)
		if math.Abs(got - test.want) > tol{
			t.Errorf("Test #%d: Wanted %f got %f\n", i, test.want, got)
		}
	}
}

func TestEvaluate(t *testing.T){
	tpcf := TwoPhaseCorrFunc{
		CorrFunc: []ReciprocalSpacePairCorrelation{
			{
				EffTemp: 0.0,
				Peaks: TriangularLattice2D(2.0, 8.0),
			},
			{
				EffTemp: 0.0,
				Peaks: TriangularLattice2D(2.0, 6.0),
			},
		},
	}

	ft := pfutil.NewFFTW([]int{16, 16})
	conc := make([]complex128, 16*16)
	for i := range conc {
		conc[i] = complex(0.1*float64(i), 0.0)
	}
	concFT := ft.FFT(conc)
	Ceff := tpcf.Evaluate(concFT, ft)
	tol := 1e-6
	for i := range Ceff {
		im := imag(Ceff[i])
		if math.Abs(im) > tol{
			t.Errorf("Test #%d: im: %f\n",i,im)
		}
	}

	// Compare resulting fourier transforms
	for i, test := range []struct{
		c float64
		w []float64
	}{
		{
			c: 0.0,
			w: []float64{0.0, 1.0},
		},
		{
			c: 1.0,
			w: []float64{1.0, 0.0},
		},
		{
			c: 0.5,
			w: []float64{0.5, 0.5},
		},
	}{
		for j := range conc{
			conc[j] = complex(test.c, 0.0)
		}
		concFT = ft.FFT(conc)
		CeffFT := tpcf.EvaluateFT(concFT, ft)
		for j := range CeffFT{
			re, im := real(CeffFT[j]), imag(CeffFT[j])
			if math.Abs(im) > tol{
				t.Errorf("Test #%d: Node %d: im: %f\n", i, j, im)
			}
			
			freq := ft.Freq(j)
			k := 2.0*math.Pi*math.Sqrt(Dot(freq, freq))
			want := test.w[0]*tpcf.CorrFunc[0].Eval(k) + test.w[1]*tpcf.CorrFunc[1].Eval(k)
			if math.Abs(re-want) > tol {
				t.Errorf("Test #%d: Node %d: want %f got %f\n", i, j, want, re)
				return
			}
		}
	}
}