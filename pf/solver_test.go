package pf

import (
	"fmt"
	"math"
	"testing"
)

func TestSolverDiffusion(t *testing.T) {
	m := NewModel()
	conc := NewField("conc", 16*16, nil)
	conc.Data[128] = 1.0
	m.AddField(conc)
	m.AddEquation("dconc/dt = LAP conc")

	solver := NewSolver(&m, []int{16, 16}, 0.1)
	solver.Solve(10, 10)

	// Check that the mass is conserved
	totalMass := 0.0
	for i := range conc.Data {
		totalMass += real(conc.Data[i])
	}

	if math.Abs(totalMass-1.0) > 1e-4 {
		t.Errorf("Mass not conserved. Initial: 1.0 Final %f", totalMass)
	}

	for i := range conc.Data {
		if real(conc.Data[i]) >= 1.0 || real(conc.Data[i]) < 0.0 {
			t.Errorf("Unexpected behavior on node %d. Value %v", i, conc.Data[i])
		}
	}
}

func TestJSONifyMonitors(t *testing.T) {
	m := NewModel()
	N := 16
	f1 := NewField("field1", N*N, nil)
	f2 := NewField("field2", N*N, nil)
	m.AddField(f1)
	m.AddField(f2)

	solver := NewSolver(&m, []int{N, N}, 0.1)
	m1 := NewPointMonitor(0, "field1")
	m2 := NewPointMonitor(1, "field2")
	m3 := NewPointMonitor(0, "field2")
	f1.Data[0] = complex(1.0, 0.0)
	m1.Add(m.Bricks)
	f2.Data[1] = complex(2.0, 0.0)
	m2.Add(m.Bricks)
	f2.Data[0] = complex(3.0, 0.0)
	m3.Add(m.Bricks)

	solver.AddMonitor(&m1)
	solver.AddMonitor(&m2)
	solver.AddMonitor(&m3)

	res := solver.JSONifyMonitors()
	strRes := fmt.Sprintf("%s", res)
	fmt.Printf(strRes)
	expect := "[{\"Data\":[1],\"Site\":0,\"Field\":\"field1\",\"Name\":\"PointMonitor\"},{\"Data\":[2],\"Site\":1,\"Field\":\"field2\",\"Name\":\"PointMonitor\"},{\"Data\":[3],\"Site\":0,\"Field\":\"field2\",\"Name\":\"PointMonitor\"}]"
	if strRes != expect {
		t.Errorf("Expected\n%s\nGot\n%s\n", expect, strRes)
	}
}

func TestSetStepperWorks(t *testing.T) {
	steppers := []string{"euler", "rk4"}
	model := NewModel()
	solver := NewSolver(&model, []int{4, 4}, 0.1)
	for _, stepper := range steppers {
		solver.SetStepper(stepper)
	}

}
