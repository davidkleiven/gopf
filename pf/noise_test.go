package pf

import (
	"math"
	"testing"
)

func TestGaussianWhiteNoise(t *testing.T) {
	model := NewModel()
	N := 16
	field := NewField("price", N*N, nil)
	noise := WhiteNoise{
		Strength: 1.0,
	}
	model.AddField(field)
	model.RegisterFunction("WHITE_NOISE", noise.Generate)
	model.AddEquation("dprice/dt = WHITE_NOISE")

	solver := NewSolver(&model, []int{N, N}, 0.1)
	solver.Solve(10, 10)
}

func TestVariance(t *testing.T) {
	num := 1000000
	data := make([]float64, num)
	noise := WhiteNoise{
		Strength: 2.0,
	}
	bricks := make(map[string]Brick)
	for i := 0; i < num; i++ {
		data[i] = real(noise.Generate(i, bricks))
	}

	mean := 0.0
	for i := range data {
		mean += data[i] / float64(len(data))
	}

	variance := 0.0
	for i := range data {
		variance += math.Pow(data[i]-mean, 2) / float64(len(data)-1)
	}
	std := math.Sqrt(variance)

	if math.Abs(std-2.0) > 0.001 {
		t.Errorf("Unexpected standard deviation. Expected 2.0 got %f", std)
	}
}

func TestConservativeNoise(t *testing.T) {
	model := NewModel()
	N := 16
	field := NewField("myfield", N*N, nil)
	model.AddField(field)

	noise := NewConservativeNoise(1.0, 2)
	dfields := noise.RequiredDerivedFields(N * N)

	model.RegisterExplicitTerm("CONSERVATIVE_NOISE", &noise, dfields)
	model.AddEquation("dmyfield/dt = CONSERVATIVE_NOISE")

	// Initialize a solver
	solver := NewSolver(&model, []int{N, N}, 0.1)
	solver.Solve(10, 100)

	// Check that the field is real
	tol := 1e-10
	for i := range field.Data {
		im := imag(field.Data[i])
		if math.Abs(im) > tol {
			t.Errorf("Imaginary field! Imag: %f\n", im)
		}
	}

	// Check that the total sum is zero
	integratedField := 0.0
	numLargerThanStd := 0
	for i := range field.Data {
		integratedField += real(field.Data[i])

		if math.Abs(real(field.Data[i])) > math.Sqrt(2.0) {
			numLargerThanStd++
		}
	}

	if math.Abs(integratedField) > tol {
		t.Errorf("Expected that the integrated field is zero. Got %f\n", integratedField)
	}

	if numLargerThanStd == 0 {
		t.Errorf("There are no nodes that are larger than the standard deviaton")
	}

}
