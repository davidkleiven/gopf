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
	solver := NewSolver(&m, []int{16, 16}, 0.1)
	m1 := NewPointMonitor(0, "field1")
	m2 := NewPointMonitor(1, "field2")
	m3 := NewPointMonitor(0, "field2")
	m1.Add(1.0)
	m2.Add(2.0)
	m3.Add(3.0)
	solver.AddMonitor(m1)
	solver.AddMonitor(m2)
	solver.AddMonitor(m3)

	res := solver.JSONifyMonitors()
	strRes := fmt.Sprintf("%s", res)
	expect := "[{\"Data\":[1],\"Site\":0,\"Field\":\"field1\"},{\"Data\":[2],\"Site\":1,\"Field\":\"field2\"},{\"Data\":[3],\"Site\":0,\"Field\":\"field2\"}]"
	if strRes != expect {
		t.Errorf("Expected\n%s\nGot\n%s\n", expect, strRes)
	}
}
