package pf

import (
	"math"
	"os"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
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

func TestSaveLoadCsv(t *testing.T) {
	model := NewModel()
	N := 8
	f1 := NewField("field1", N*N, nil)
	f2 := NewField("field2", N*N, nil)
	for i := 0; i < N*N; i++ {
		f1.Data[i] = complex(float64(i-i*i), 0.0)
		f2.Data[i] = complex(float64(i), 0.0)
	}
	model.AddField(f1)
	model.AddField(f2)

	solver := NewSolver(&model, []int{N, N}, 0.1)

	csvIO := CsvIO{
		Prefix:     "my_csv_file",
		DomainSize: []int{N, N},
	}

	csvIO.SaveFields(solver, 0)

	loaded := LoadCSV("my_csv_file_0.csv")
	tol := 1e-8
	fields := []Field{f1, f2}
	for i := range loaded {
		if !pfutil.CmplxEqualApprox(loaded[i].Data, fields[i].Data, tol) {
			t.Errorf("Expected\n%v\nGot\n%v\n", fields[i].Data, loaded[i].Data)
		}
	}
	os.Remove("my_csv_file_0.csv")
}
