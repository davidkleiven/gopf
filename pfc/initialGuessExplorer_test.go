package pfc

import (
	"gonum.org/v1/gonum/floats"
	"testing"
)

func TestHyperOctant(t *testing.T) {
	hyper := HyperOctantExplorer{
		X0: []float64{1.0, 2.0, 3.0},
	}

	expect := [][]float64{
		[]float64{1.0, 2.0, 3.0},
		[]float64{-1.0, 2.0, 3.0},
		[]float64{1.0, -2.0, 3.0},
		[]float64{1.0, 2.0, -3.0},
		[]float64{-1.0, 2.0, -3.0},
		[]float64{1.0, -2.0, -3.0},
		[]float64{-1.0, -2.0, -3.0},
		[]float64{-1.0, -2.0, 3.0},
	}

	found := make([]bool, len(expect))
	counter := 0

	for x0 := hyper.Next(); x0 != nil; x0 = hyper.Next() {
		for j := range expect {
			if floats.EqualApprox(x0, expect[j], 1e-10) {
				if found[j] {
					t.Errorf("Starting point %v was encountered before\n", x0)
					return
				}
				found[j] = true
			}
		}
		counter++
	}

	if counter != len(expect) {
		t.Errorf("Expected %d starting points, got %d\n", len(expect), counter)
	}

	for j := range found {
		if !found[j] {
			t.Errorf("Combination %d never occured", j)
		}
	}
}
