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
