package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

// SinX constructs a 2D dataset that is sinusoidal in x-direction
func SinX(nx, ny int) ([]float64, []float64) {
	data := make([]float64, nx*ny)
	gradSq := make([]float64, nx*ny)
	for i := range data {
		c := i % nx
		x := float64(c) / float64(nx)
		data[i] = math.Sin(2.0 * math.Pi * x)
		gradSq[i] = math.Pow(2.0*math.Pi*math.Cos(2.0*math.Pi*x)/float64(nx), 2.0)
	}
	return data, gradSq
}

// Poly4 returns the polynomial P(x, y) = g(x)*g(y), where g(x) = 16*(x^2 - 2*x^3 + x^4)
func Poly4(nx, ny int) ([]float64, []float64) {
	data := make([]float64, nx*ny)
	gradSq := make([]float64, nx*ny)
	for i := range data {
		r := i / nx
		c := i % nx
		x := float64(c) / float64(nx)
		y := float64(r) / float64(ny)
		vx := 16.0 * (x*x - 2*x*x*x + x*x*x*x)
		vy := 16.0 * (y*y - 2*y*y*y + y*y*y*y)
		data[i] = vx * vy
		ddx := 16.0 * (2*x - 6*x*x + 4*x*x*x) * vy / float64(nx)
		ddy := vx * 16.0 * (2*y - 6*y*y + 4*y*y*y) / float64(ny)
		gradSq[i] = ddx*ddx + ddy*ddy
	}
	return data, gradSq
}

func TestSquareGradient(t *testing.T) {
	for i, test := range []struct {
		F   func(nx, ny int) ([]float64, []float64)
		N   int
		tol float64
	}{
		{
			F:   SinX,
			N:   16,
			tol: 1e-10,
		},
		{
			F:   Poly4,
			N:   32,
			tol: 1e-5,
		},
	} {
		bricks := make(map[string]Brick)
		N := test.N
		data, gradSq := test.F(N, N)

		cData := sfft.ToComplex(data)

		grad := NewSquareGradient("height", []int{N, N})
		field := NewField("height", N*N, cData)
		grad.FT.FFT(field.Data)
		bricks["height"] = field

		res := make([]complex128, len(data))
		function := grad.Construct(bricks)
		function(grad.FT.Freq, 0.0, res)
		grad.FT.IFFT(res)
		for j := range res {
			re := real(res[j]) / float64(N*N)
			im := imag(res[j]) / float64(N*N)
			if math.Abs(re-gradSq[j]) > test.tol || math.Abs(im) > test.tol {
				t.Errorf("Test #%d. Expected (%f, 0), got (%f, %f)", i, gradSq[j], re, im)
			}
		}
	}
}

func TestSquareGradWithSolver(t *testing.T) {
	N := 16
	model := NewModel()
	field1 := NewField("field1", N*N, nil)
	field2 := NewField("field2", N*N, nil)
	data, gradSq := SinX(N, N)
	for i := range field2.Data {
		field2.Data[i] = complex(data[i], 0.0)
	}
	grad := NewSquareGradient("field2", []int{N, N})
	zero := NewScalar("ZERO", complex(0.0, 0.0))

	model.AddField(field1)
	model.AddField(field2)
	model.AddScalar(zero)
	model.RegisterUserDefinedTerm("GRAD_SQ_f2", &grad, nil)

	model.AddEquation("dfield1/dt = GRAD_SQ_f2")
	model.AddEquation("dfield2/dt = ZERO*field1")

	dt := 0.1
	solver := NewSolver(&model, []int{N, N}, dt)

	nsteps := 10
	solver.Solve(1, nsteps)

	// Expect no change in field2
	tol := 1e-10
	for i := range data {
		re := real(model.Bricks["field2"].Get(i))
		im := imag(model.Bricks["field2"].Get(i))

		if math.Abs(re-data[i]) > tol || math.Abs(im) > tol {
			t.Errorf("Unexpected value at node %d: Expected (%f, 0) got (%f, %f)\n", i, data[i], re, im)
		}

		re = real(model.Bricks["field1"].Get(i))
		im = imag(model.Bricks["field1"].Get(i))

		expect := dt * float64(nsteps) * gradSq[i]
		if math.Abs(re-expect) > tol || math.Abs(im) > tol {
			t.Errorf("Unexpected field 1 at node %d: Expected (%f, 0) got (%f, %f)", i, expect, re, im)
		}
	}
}
