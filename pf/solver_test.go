package pf

import (
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
