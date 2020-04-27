package pf

import (
	"math"
	"testing"
)

func TestImplicitEuler(t *testing.T) {
	N := 8
	for i, test := range []struct {
		Fields   []string
		Eqns     []string
		Init     []float64
		Dt       float64
		Nsteps   int
		Solution func(t float64) []float64
		Tol      float64
	}{
		// Test equation with only a linear term
		{
			Fields: []string{"conc"},
			Eqns:   []string{"dconc/dt = m1*conc"},
			Init:   []float64{1.0},
			Dt:     0.01,
			Nsteps: 100,
			Solution: func(t float64) []float64 {
				return []float64{math.Exp(-t)}
			},
			Tol: 0.005,
		},
		// Test equation with only a non-linear term
		{
			Fields: []string{"conc"},
			Eqns:   []string{"dconc/dt = m1*conc^2"},
			Init:   []float64{1.0},
			Dt:     0.01,
			Nsteps: 100,
			Solution: func(t float64) []float64 {
				return []float64{1.0 / (1.0 + t)}
			},
			Tol: 0.005,
		},
		// Test equation with both linear and non-linear term
		{
			Fields: []string{"conc"},
			Eqns:   []string{"dconc/dt = conc+m1*conc^2"},
			Init:   []float64{0.5},
			Dt:     0.01,
			Nsteps: 100,
			Solution: func(t float64) []float64 {
				return []float64{math.Exp(t) / (1.0 + math.Exp(t))}
			},
			Tol: 0.005,
		},
		// Test equation with both linear and non-linear term and two coupled fields
		{
			Fields: []string{"conc1", "conc2"},
			Eqns:   []string{"dconc1/dt = m1*conc1*conc2", "dconc2/dt = m1*conc2"},
			Init:   []float64{1.0, 1.0},
			Dt:     0.01,
			Nsteps: 100,
			Solution: func(t float64) []float64 {
				return []float64{math.Exp(math.Exp(-t) - 1.0), math.Exp(-t)}
			},
			Tol: 0.005,
		},
	} {
		model := NewModel()
		domainSize := []int{N, N}
		for j, name := range test.Fields {
			field := NewField(name, N*N, nil)
			for k := range field.Data {
				field.Data[k] = complex(test.Init[j], 0.0)
			}
			model.AddField(field)
		}
		model.AddScalar(Scalar{
			Name:  "m1",
			Value: complex(-1.0, 0.0),
		})

		for _, eq := range test.Eqns {
			model.AddEquation(eq)
		}

		solver := ImplicitEuler{
			FT: NewFFTW(domainSize),
			Dt: test.Dt,
		}

		model.Init()
		model.Summarize()
		for step := 0; step < test.Nsteps; step++ {
			solver.Step(&model)
		}

		expect := test.Solution(test.Dt * float64(test.Nsteps))
		for j, f := range model.Fields {
			for k := range f.Data {
				re := real(f.Data[k])
				im := imag(f.Data[k])

				if math.Abs(re-expect[j]) > test.Tol || math.Abs(im) > test.Tol {
					t.Errorf("Test #%d field no. %d: Expected %f got (%f, %f)\n", i, j, expect[j], re, im)
				}
			}
		}
	}
}

func TestField2VecRoundTrip(t *testing.T) {
	N := 16
	fields := []Field{
		NewField("f1", N*N, nil),
		NewField("f2", N*N, nil),
	}

	origFields := make([]Field, 2)

	for i := range fields {
		for j := range fields[i].Data {
			fields[i].Data[j] = complex(float64(i*j), 0.0)
		}
	}

	// Fourier transform the fields
	ft := NewFFTW([]int{N, N})
	ft.FFT(fields[0].Data)
	ft.FFT(fields[1].Data)

	for i := 0; i < len(origFields); i++ {
		origFields[i] = fields[i].Copy()
	}
	out := make([]float64, 2*N*N)
	ie := ImplicitEuler{
		FT: ft,
	}

	ie.fields2vec(fields, out)

	// Clear all fields
	for i := range fields {
		Clear(fields[i].Data)
	}

	// Confirm that all fields values are now zero
	tol := 1e-10
	for i := range fields {
		for j := range fields[i].Data {
			re := real(fields[i].Data[j])
			im := imag(fields[i].Data[j])
			if math.Abs(re) > tol || math.Abs(im) > tol {
				t.Errorf("Clear did not set all values to zero")
			}
		}
	}

	ie.vec2fields(out, fields)

	// Confirm that the fields now matches
	for i := range fields {
		if !CmplxEqualApprox(fields[i].Data, origFields[i].Data, tol) {
			t.Errorf("Expected %v\ngot\n%v\n", origFields[i].Data, fields[i].Data)
		}
	}
}

func TestDissipatingHeatEquation(t *testing.T) {
	N := 128
	field := NewField("temperature", N*N, nil)

	// Initialize the field
	for i := range field.Data {
		ix := i / N
		iy := i % N
		x := float64(ix) / float64(N)
		y := float64(iy) / float64(N)
		value := math.Sin(2.0*x*math.Pi) * math.Sin(2.0*y*math.Pi)
		field.Data[i] = complex(value, 0.0)
	}

	model := NewModel()
	gamma := 0.2

	// For the sake of testing a linear dissipation is added via a function
	dissipation := func(i int, bricks map[string]Brick) complex128 {
		return -complex(gamma, 0.0) * bricks["temperature"].Get(i)
	}

	model.AddField(field)
	model.RegisterFunction("DISSIPATE", dissipation)
	model.AddEquation("dtemperature/dt = LAP temperature + DISSIPATE")

	dt := 0.005
	solver := NewSolver(&model, []int{N, N}, dt)
	stepper := ImplicitEuler{
		Dt: dt,
		FT: NewFFTW([]int{N, N}),
	}
	solver.Stepper = &stepper
	nsteps := 10
	solver.Solve(1, nsteps)

	analytical := func(x, y, t float64) float64 {
		L := float64(N)
		return math.Exp(-(4.0*math.Pi/(L*L)+gamma)*t) * math.Sin(2.0*math.Pi*x) * math.Sin(2.0*math.Pi*y)
	}

	finalTime := float64(nsteps) * dt
	tol := 1e-3
	for i := range field.Data {
		ix := i / N
		iy := i % N
		x := float64(ix) / float64(N)
		y := float64(iy) / float64(N)
		value := analytical(y, x, finalTime)

		re := real(field.Data[i])
		im := imag(field.Data[i])

		if math.Abs(re-value) > tol || math.Abs(im) > tol {
			t.Errorf("Expected %f got (%f, %f) at position (%f, %f)\n", value, re, im, x, y)
		}
	}
}
