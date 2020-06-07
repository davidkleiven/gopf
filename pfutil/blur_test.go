package pfutil

import (
	"fmt"
	"math"
	"testing"
)

func TestBlur(t *testing.T) {
	N := 64
	data := make([]float64, N*N)
	integral := 0.0
	for i := range data {
		x := Pos([]int{N, N}, i)
		if x[0] > N/2 {
			data[i] = 1.0
		}
		integral += data[i]
	}
	dataOrig := make([]float64, len(data))
	copy(dataOrig, data)

	Blur(data, []int{N, N}, &BoxKernel{Width: 2})

	// Check that the integral does not change
	newIntegral := 0.0
	for i := range data {
		newIntegral += data[i]
	}

	if math.Abs(integral-newIntegral) > 1e-10 {
		t.Errorf("Integral changed. Expected %f got %f\n", integral, newIntegral)
	}

	// Check that all values that where 1 decreased and all value that were
	// zero increases
	for i := range data {
		if dataOrig[i] > 0.9 {
			if data[i] > 1.0 {
				t.Errorf("Expected a value smaller than or equal to 1. Got %f\n", data[i])
			}
		} else {
			if data[i] < 0.0 {
				t.Errorf("Expected value a greater than or equal to 0. Got %f\n", data[i])
			}
		}
	}
	fmt.Printf("%v\n", data)
}
