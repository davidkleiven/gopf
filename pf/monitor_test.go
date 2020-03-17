package pf

import (
	"testing"

	"gonum.org/v1/gonum/floats"
)

func TestMonitor(t *testing.T) {
	monitor := NewPointMonitor(0, "myfield")
	field := NewField("myfield", 8, nil)
	bricks := make(map[string]Brick)
	bricks["myfield"] = field
	monitor.Add(bricks)
	field.Data[0] = complex(1.0, 0.0)
	monitor.Add(bricks)
	field.Data[0] = complex(2.0, 0.0)
	monitor.Add(bricks)

	expect := []float64{0.0, 1.0, 2.0}

	if !floats.EqualApprox(expect, monitor.Data, 1e-10) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expect, monitor.Data)
	}
}
