package pf

import (
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestMonitor(t *testing.T) {
	monitor := NewPointMonitor(0, "myfield")
	monitor.Add(1.0)
	monitor.Add(2.0)
	monitor.Add(3.0)

	expect := []float64{1.0, 2.0, 3.0}

	if !floats.EqualApprox(expect, monitor.Data, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, monitor.Data)
	}
}
