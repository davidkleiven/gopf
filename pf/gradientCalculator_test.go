package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
	"github.com/davidkleiven/gosfft/sfft"
)

func TestGradientCalculator(t *testing.T) {
	N := 16
	data := make([]complex128, N*N)
	expect := make([]float64, N*N)
	for i := range data {
		x := float64(i%N) / float64(N)
		data[i] = complex(x*x-2*x*x*x+x*x*x*x, 0.0)
		v := 2.0*x - 6.0*x*x + 4.0*x*x*x
		expect[i] = v / float64(N)
	}

	ft := sfft.NewFFT2(N, N)
	grad := GradientCalculator{
		FT:   ft,
		Comp: 1,
	}

	got := make([]complex128, N*N)
	grad.Calculate(data, got)
	tol := 1e-4
	for i := range got {
		re := real(got[i])
		im := imag(got[i])
		if math.Abs(re-expect[i]) > tol || math.Abs(im) > tol {
			diff := re - expect[i]
			t.Errorf("Expected (%f, 0) got (%f, %f). Real part diff %f\n", expect[i], re, im, diff)
		}
	}
}

func GaussianProfile(x, y, sigma float64) float64 {
	rSq := x*x + y*y
	return math.Exp(-0.5 * rSq / (sigma * sigma))
}

func FillData(nx, ny int) ([]float64, []float64) {
	data := make([]float64, nx*ny)
	dg := make([]float64, nx*ny)
	sigma := 1.0 / 10.0
	for i := 0; i < nx*ny; i++ {
		x := float64(i%nx) / float64(nx)
		y := float64(i/nx) / float64(ny)
		x -= 0.5
		y -= 0.5
		data[i] = GaussianProfile(x, y, sigma)
		dg[i] = ((2.0*x*x+2.0*y*y)/(sigma*sigma) - 2.0) * data[i] * data[i] / (sigma * sigma)
	}
	return data, dg
}

func TestDivGrad(t *testing.T) {
	divGrad := DivGrad{
		Field: "myfield",
		F: func(i int, bricks map[string]Brick) complex128 {
			return bricks["myfield"].Get(i)
		},
	}

	got := divGrad.FuncName()
	want := "DivGrad_myfield_Func"
	if got != want {
		t.Errorf("Expected: %s got %s\n", want, got)
	}

	got = divGrad.GradName(1)
	want = "GRAD_myfield_1"

	if got != want {
		t.Errorf("Expected: %s got %s\n", want, got)
	}

	// Compare against known solution
	N := 64
	model := NewModel()
	field := NewField("myfield", N*N, nil)
	data, dg := FillData(N, N)
	for i := range data {
		field.Data[i] = complex(data[i], 0.0)
	}
	model.AddField(field)

	ft := sfft.NewFFT2(N, N)
	divGrad.PrepareModel(N*N, &model, ft)
	model.Init()

	rhs := divGrad.Construct(model.Bricks)
	for i := range model.Fields {
		ft.FFT(model.Fields[i].Data)
	}
	for i := range model.DerivedFields {
		ft.FFT(model.DerivedFields[i].Data)
	}

	res := make([]complex128, len(data))
	rhs(ft.Freq, 0.0, res)
	ft.IFFT(res)
	pfutil.DivRealScalar(res, float64(len(res)))

	tol := 1e-3
	mismatch := false
	maxdiff := 0.0
	maxrelDiff := 0.0
	for i := range dg {
		re := real(res[i]) * float64(N*N)
		expect := dg[i]
		atolOK := math.Abs(re-expect) < tol
		rtolOK := math.Abs(re-expect) < expect*tol
		if !atolOK && !rtolOK {
			mismatch = true
			if math.Abs(re-expect) > maxdiff {
				maxdiff = math.Abs(re - expect)
				maxrelDiff = maxdiff / math.Abs(expect+tol)
			}
		}
	}

	if mismatch {
		t.Errorf("Mismatch. Maximum difference: %f (%f %%)\n", maxdiff, 100.0*maxrelDiff)
	}

	// Make sure that the data is real
	isReal := true
	tol = 1e-10
	maxIm := 0.0
	for i := range res {
		im := math.Abs(imag(res[i]))
		if im > tol {
			isReal = false
			if im > maxIm {
				maxIm = im
			}
		}
	}

	if !isReal {
		t.Errorf("DivGrad: Result is not real. Max. imaginary part: %e\n", maxIm)
	}
}

func TestWeightedLaplacian(t *testing.T) {
	N := 16
	prefactor := NewField("prefactor", N*N, nil)
	field := NewField("field", N*N, nil)
	expect := make([]complex128, N*N)
	twoPi := 2.0 * math.Pi
	for i := range prefactor.Data {
		pos := pfutil.Pos([]int{N, N}, i)
		x := float64(pos[0]) / float64(N)
		prefactor.Data[i] = complex(math.Sin(twoPi*x), 0.0)
		field.Data[i] = complex(math.Cos(twoPi*x), 0.0)
		expect[i] = complex(-math.Pow(twoPi, 2)*math.Sin(twoPi*x)*math.Cos(twoPi*x)/float64(N*N), 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["prefactor"] = prefactor
	bricks["field"] = field

	ft := pfutil.NewFFTW([]int{N, N})

	// Fourier transform the fields
	ft.FFT(field.Data)
	ft.FFT(prefactor.Data)

	wl := WeightedLaplacian{
		Field:     "field",
		PreFactor: "prefactor",
		FT:        ft,
	}

	function := wl.Construct(bricks)
	result := make([]complex128, N*N)
	function(ft.Freq, 0.0, result)

	ft.IFFT(result)
	pfutil.DivRealScalar(result, float64(len(result)))

	if !pfutil.CmplxEqualApprox(expect, result, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, result)
	}
}
