package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gopf/elasticity"
	"gonum.org/v1/gonum/mat"
)

func TestIndicatorDeriv(t *testing.T) {
	dx := 0.01
	tol := 1e-4
	for i := 0; i < 100; i++ {
		x := dx * float64(i)
		h1 := Indicator(x - dx/2.0)
		h2 := Indicator(x + dx/2.0)
		dhdx := (h2 - h1) / dx
		expect := IndicatorDeriv(dx * float64(i))
		if math.Abs(dhdx-expect) > tol {
			t.Errorf("Derivative does not match. Expected %f got %f", expect, dhdx)
		}
	}
}

func TestHomogeneousRHS(t *testing.T) {
	N := 16
	matProp := elasticity.Isotropic(60.0, 0.3)
	eps := 0.01
	misfit := mat.NewDense(3, 3, []float64{eps, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0})

	ux := make([]complex128, N*N)
	exx := make([]float64, N*N)
	for i := range ux {
		row := i / N
		x := float64(row) / float64(N)
		ux[i] = complex(math.Pow(x*(1.0-x), 2), 0.0)
		exx[i] = 2 * x * (1 - x) * (1 - 2*x) / float64(N)
	}

	homogenous := NewHomogeneousModolus("x", []int{N, N}, matProp, misfit)
	homogenous.FT.FFT(ux)

	homogenous.Disps = func(force [][]complex128, freq elasticity.Frequency, matProp elasticity.Rank4Tensor) [][]complex128 {
		res := make([][]complex128, len(force))
		for i := range res {
			res[i] = make([]complex128, 2)
			res[i][0] = ux[i]
		}
		return res
	}

	bricks := make(map[string]Brick)
	f := NewField("elasticity", N*N, nil)

	eta := 0.5
	for i := range f.Data {
		f.Data[i] = complex(eta, 0.0)
		homogenous.Field[i] = real(f.Data[i])
	}
	bricks["elasticity"] = f

	expect := make([]float64, N*N)
	for i := range expect {
		C := matProp.At(0, 0, 0, 0)
		H := Indicator(eta)
		dHdx := IndicatorDeriv(eta)
		value := C * eps * (exx[i] - eps*H) * dHdx
		expect[i] = value

	}

	function := homogenous.Construct(bricks)
	res := make([]complex128, N*N)
	function(homogenous.FT.Freq, 0.0, res)
	homogenous.FT.IFFT(res)
	DivRealScalar(res, float64(len(res)))

	tol := 1e-4
	for i := range res {
		re := real(res[i])
		im := imag(res[i])
		if math.Abs(re-expect[i]) > tol || math.Abs(im) > tol {
			t.Errorf("Unexpected value. Expected (%f, 0) got (%f, %f)\n", expect[i], re, im)
		}
	}
}
