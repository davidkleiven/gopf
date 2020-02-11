package pf

import (
	"math"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	model := NewModel()
	N := 8
	field := NewField("conc", N*N, nil)
	for i := 0; i < N*N; i++ {
		field.Data[i] = complex(float64(i), 0.0)
	}

	model.AddField(field)

	solver := NewSolver(&model, []int{N, N}, 0.1)
	writer := NewFloat64IO("myfile")
	writer.SaveFields(solver, 0)

	data := LoadFloat64("myfile_conc_0.bin")

	tol := 1e-10
	for i := range data {
		re := real(field.Data[i])
		if math.Abs(data[i]-re) > tol {
			t.Errorf("Expected %f got %f\n", data[i], re)
		}
	}
}
