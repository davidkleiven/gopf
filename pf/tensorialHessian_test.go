package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

func GetData(N int) map[string][]float64 {
	H := make([]float64, N*N)
	dHdx2 := make([]float64, N*N)
	dHdxdy := make([]float64, N*N)
	dHdy2 := make([]float64, N*N)

	for i := range H {
		y := float64(i%N) / float64(N)
		x := float64(i/N) / float64(N)

		px := 16.0 * (x*x - 2.0*x*x*x + x*x*x*x)
		py := 16.0 * (y*y - 2.0*y*y*y + y*y*y*y)
		H[i] = px * py
		dpxdx2 := 16.0 * (2.0 - 12.0*x + 12*x*x)
		dpydy2 := 16.0 * (2.0 - 12.0*y + 12.0*y*y)
		dpxdx := 16.0 * (2.0*x - 6.0*x*x + 4.0*x*x*x)
		dpydy := 16.0 * (2.0*y - 6.0*y*y + 4.0*y*y*y)

		dHdx2[i] = dpxdx2 * py / float64(N*N)
		dHdy2[i] = px * dpydy2 / float64(N*N)
		dHdxdy[i] = dpxdx * dpydy / float64(N*N)
	}

	res := make(map[string][]float64)
	res["data"] = H
	res["dx2"] = dHdx2
	res["dy2"] = dHdy2
	res["dxdy"] = dHdxdy
	return res
}

func TestTensorialHessian(t *testing.T) {
	N := 64
	data := GetData(N)

	ft := sfft.NewFFT2(N, N)

	cmplxData := make([]complex128, N*N)
	for i := range data["data"] {
		cmplxData[i] = complex(data["data"][i], 0.0)
	}

	ft.FFT(cmplxData)

	field := NewField("myfield", N*N, cmplxData)
	bricks := make(map[string]Brick)
	bricks["myfield"] = field

	for i, test := range []struct {
		K   []float64
		Res string
		tol float64
	}{
		{
			K:   []float64{1.0, 0.0, 0.0, 0.0},
			Res: "dx2",
			tol: 1e-3,
		},
		{
			K:   []float64{0.0, 0.0, 0.0, 1.0},
			Res: "dy2",
			tol: 1e-3,
		},
		{
			K:   []float64{0.0, 0.5, 0.5, 0.0},
			Res: "dxdy",
			tol: 1e-6,
		},
	} {
		hessian := TensorialHessian{
			K:     test.K,
			Field: "myfield",
		}

		function := hessian.Construct(bricks)
		res := make([]complex128, N*N)
		function(ft.Freq, 0.0, res)
		ft.IFFT(res)
		DivRealScalar(res, float64(len(res)))

		for j := range res {
			re := real(res[j])
			im := imag(res[j])

			want := data[test.Res][j]
			if math.Abs(re-want) > test.tol || math.Abs(im) > test.tol {
				t.Errorf("Test #%d: Want (%f, 0) got (%f, %f)\n", i, want, re, im)
			}
		}
	}
}
