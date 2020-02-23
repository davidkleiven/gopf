package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

const sigma = 1.0 / 10.0

func FillGaussianProfile(nx, ny int) []complex128 {
	data := make([]complex128, nx*ny)
	for i := range data {
		x := float64(i%nx) / float64(nx)
		y := float64(i/nx) / float64(ny)
		x -= 0.5
		y -= 0.5
		rSq := x*x + y*y
		data[i] = complex(math.Exp(-0.5*rSq/(sigma*sigma)), 0.0)
	}
	return data
}

func ConstVelocityXDir(nx, ny int) ([]complex128, []complex128) {
	vx := make([]complex128, nx*ny)
	vy := make([]complex128, nx*ny)
	for i := range vx {
		vx[i] = complex(1.0, 0.0)
	}
	return vx, vy
}

func ConstVelocityYDir(nx, ny int) ([]complex128, []complex128) {
	vx := make([]complex128, nx*ny)
	vy := make([]complex128, nx*ny)
	for i := range vx {
		vy[i] = complex(1.0, 0.0)
	}
	return vx, vy
}

func LinearVelocityX(nx, ny int) ([]complex128, []complex128) {
	vx := make([]complex128, nx*ny)
	vy := make([]complex128, nx*ny)
	for i := range vx {
		x := float64(i%nx) / float64(nx)
		x -= 0.5
		vx[i] = complex(x, 0.0)
	}
	return vx, vy
}

func LinearVelocityVxY(nx, ny int) ([]complex128, []complex128) {
	vx := make([]complex128, nx*ny)
	vy := make([]complex128, nx*ny)
	for i := range vx {
		y := float64(i%nx) / float64(nx)
		y -= 0.5
		vx[i] = complex(y, 0.0)
	}
	return vx, vy
}

func ExpectGaussConstVel(nx, ny, dir int) []float64 {
	data := make([]float64, nx*ny)
	for i := range data {
		x := float64(i%nx) / float64(nx)
		y := float64(i/nx) / float64(ny)
		x -= 0.5
		y -= 0.5
		rSq := x*x + y*y
		if dir == 0 {
			data[i] = x * math.Exp(-0.5*rSq/(sigma*sigma)) / (sigma * sigma)
		} else {
			data[i] = y * math.Exp(-0.5*rSq/(sigma*sigma)) / (sigma * sigma)
		}
	}
	return data
}

func ExpectGaussLinVelVxY(nx, ny int) []float64 {
	data := make([]float64, nx*ny)
	for i := range data {
		x := float64(i%nx) / float64(nx)
		y := float64(i/nx) / float64(ny)
		x -= 0.5
		y -= 0.5
		rSq := x*x + y*y
		data[i] = y * x * math.Exp(-0.5*rSq/(sigma*sigma)) / (sigma * sigma)
	}
	return data
}

func TestAdvection(t *testing.T) {

	for i, test := range []struct {
		FieldInit    func(nx, ny int) []complex128
		VelocityInit func(nx, ny int) ([]complex128, []complex128)
		Expect       func(nx, ny int) []float64
	}{
		{
			FieldInit:    FillGaussianProfile,
			VelocityInit: ConstVelocityXDir,
			Expect: func(nx, ny int) []float64 {
				return ExpectGaussConstVel(nx, ny, 1)
			},
		},
		{
			FieldInit:    FillGaussianProfile,
			VelocityInit: ConstVelocityYDir,
			Expect: func(nx, ny int) []float64 {
				return ExpectGaussConstVel(nx, ny, 0)
			},
		},
		{
			FieldInit:    FillGaussianProfile,
			VelocityInit: LinearVelocityX,
			Expect:       ExpectGaussLinVelVxY,
		},
	} {
		N := 64
		model := NewModel()
		fieldData := test.FieldInit(N, N)
		field := NewField("conc", N*N, fieldData)
		vx, vy := test.VelocityInit(N, N)
		vxField := NewField("vx", N*N, vx)
		vyField := NewField("vy", N*N, vy)

		model.AddField(field)
		model.AddField(vxField)
		model.AddField(vyField)

		advection := Advection{
			Field:          "conc",
			VelocityFields: []string{"vx", "vy"},
		}
		ft := sfft.NewFFT2(N, N)
		advection.PrepareModel(N*N, &model, ft)
		model.Init()

		rhsFunc := advection.Construct(model.Bricks)
		res := make([]complex128, N*N)
		rhsFunc(ft.Freq, 0.0, res)
		expect := test.Expect(N, N)

		maxDiff := 0.0
		relDiff := 0.0
		match := true
		tol := 1e-3
		for j := range res {
			re := real(res[j]) * float64(N)
			im := imag(res[j]) * float64(N)

			diff := math.Abs(re - expect[j])
			if diff > tol || math.Abs(im) > tol {
				match = false
			}

			if diff > maxDiff {
				maxDiff = diff
				relDiff = diff / (math.Abs(expect[j]) + tol)
			}
		}

		if !match {
			t.Errorf("Test #%d: Max. diff %e (%f %%)\n", i, maxDiff, 100.0*relDiff)
		}

	}
}

func TestAdvectionPanics(t *testing.T) {
	model := NewModel()
	advection := Advection{
		Field:          "conc",
		VelocityFields: []string{"vx"},
	}

	ft := sfft.NewFFT2(8, 8)

	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("Test #1: Expected panic")
			}
		}()
		advection.PrepareModel(64, &model, ft)
	}()

	field := NewField("conc", 64, nil)
	vx := NewField("vx", 64, nil)
	vy := NewField("vy", 64, nil)
	model.AddField(field)
	model.AddField(vx)
	model.AddField(vy)

	// Should panic because wrong number of velocity fields
	func() {
		defer func() {
			if recover() == nil {
				t.Errorf("Test #2: Expected panic")
			}
		}()
		advection.PrepareModel(64, &model, ft)
	}()

	advection.VelocityFields = []string{"vx", "vy"}

	// Should not panic
	advection.PrepareModel(64, &model, ft)
}
