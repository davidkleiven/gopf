package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gopf/pfc"
)

func TestPairCorrelationTerm(t *testing.T) {
	pair := PairCorrlationTerm{
		PairCorrFunc: pfc.ReciprocalSpacePairCorrelation{
			EffTemp: 0.0,
			Peaks: []pfc.Peak{
				pfc.Peak{
					PlaneDensity: 1,
					Location:     1.0,
					Width:        100.0,
					NumPlanes:    1,
				},
			},
		},
		Field:     "myfield",
		Prefactor: 1.0,
	}

	N := 16
	field := NewField("myfield", N*N, nil)

	// Insert fourier transformed fields
	for i := range field.Data {
		field.Data[i] = complex(0.1*float64(i), 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["myfield"] = field
	function := pair.Construct(bricks)
	res := make([]complex128, N*N)

	freq := func(i int) []float64 {
		return []float64{float64(i), float64(2 * i)}
	}
	function(freq, 0.0, res)

	for i := range field.Data {
		f := freq(i)
		fRad := 2.0 * math.Pi * math.Sqrt(Dot(f, f))
		wSq := math.Pow(pair.PairCorrFunc.Peaks[0].Width, 2)
		factor := math.Exp(-0.5 * (fRad - 1.0) * (fRad - 1.0) / wSq)
		expect := -factor
		re := real(res[i])

		if math.Abs(re-expect) > 1e-10 {
			t.Errorf("Expected %f got %f\n", expect, re)
		}
	}
}

func TestPairCorrelationGetEnergy(t *testing.T) {
	wavenumber := 2.0 * math.Pi / 4
	pair := PairCorrlationTerm{
		PairCorrFunc: pfc.ReciprocalSpacePairCorrelation{
			EffTemp: 0.0,
			Peaks: []pfc.Peak{
				pfc.Peak{
					PlaneDensity: 1,
					Location:     wavenumber,
					Width:        100.0,
					NumPlanes:    1,
				},
			},
		},
		Field:     "myfield",
		Prefactor: 1.0,
	}

	N := 16
	field := NewField("myfield", N*N, nil)

	// Insert a cosine
	for i := range field.Data {
		x := float64(i % N)
		field.Data[i] = complex(math.Cos(wavenumber*x), 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["myfield"] = field
	ft := NewFFTW([]int{N, N})
	energy := pair.GetEnergy(bricks, ft, []int{N, N})
	expect := -0.5 * float64(N*N) * 0.5
	tol := 1e-10
	if math.Abs(energy-expect) > tol {
		t.Errorf("Expected energy %f. Got %f\n", expect, energy)
	}
}

func TestIdealMixTerm(t *testing.T) {
	term := IdealMixtureTerm{
		IdealMix:  pfc.IdealMix{C3: 1.0, C4: 1.0},
		Field:     "eta",
		Prefactor: 1.0,
	}

	N := 16
	data := make([]complex128, N*N)
	for i := range data {
		data[i] = complex(0.1*float64(i), 0.0)
	}

	field := NewField("eta", N*N, data)
	bricks := make(map[string]Brick)
	bricks["eta"] = field
	tol := 1e-10

	for i := range data {
		val := term.Eval(i, bricks)
		expect := term.IdealMix.Deriv(real(data[i]))
		re := real(val)
		if math.Abs(re-expect) > tol {
			t.Errorf("Expected %f. Got %f\n", expect, re)
		}
	}

	// Test get energy
	for i := 0; i < N*N; i++ {
		field.Data[i] = complex(1.0, 0.0)
	}
	bricks["eta"] = field
	energy := term.GetEnergy(bricks, N*N)
	expect := 5.0 * float64(N*N) / 12.0
	if math.Abs(energy-expect) > tol {
		t.Errorf("Expected energy %f. Got %f\n", expect, energy)
	}
}

func TestPeakWidthConsistency(t *testing.T) {
	// This is a quite complicated test, that checks that increase in energy when
	// compressing the cell, agrees with the bulk modulus when the widths are tuned
	// such that the bulk modulus match
	a := 16.0
	cell := pfc.SC2D(a)
	reciprocal := cell.Reciprocal()
	modeSolver := pfc.ModeSolver{
		IdealMix: pfc.IdealMix{C3: 1.0, C4: 1.0},
		Miller: []pfc.Miller{
			pfc.Miller{H: 1, K: 1},
			pfc.Miller{H: 2, K: 0},
		},
		ReciprocalCell: reciprocal,
	}

	// Define amplitudes
	modeResult := modeSolver.Solve(nil)
	A11 := modeResult.Amplitudes[0]
	A20 := modeResult.Amplitudes[1]
	N := 64
	field := NewField("density", N*N, nil)
	for i := range field.Data {
		x := float64(i % N)
		y := float64(i / N)
		sx := x/a + 0.5/a
		sy := y/a + 0.5/a
		density11 := modeSolver.ModeDensity(0, []float64{sx, sy})
		density20 := modeSolver.ModeDensity(1, []float64{sx, sy})
		field.Data[i] = complex(A11*density11+A20*density20, 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["density"] = field

	wavenumber1 := 2.0 * math.Pi * math.Sqrt(2.0) / a
	wavenumber2 := 2.0 * math.Pi * 2.0 / a
	term := PairCorrlationTerm{
		PairCorrFunc: pfc.ReciprocalSpacePairCorrelation{
			EffTemp: 0.0,
			Peaks: []pfc.Peak{
				pfc.Peak{
					PlaneDensity: 1.0,
					Location:     wavenumber1,
					Width:        0.01,
					NumPlanes:    pfc.NumEquivalent2D(pfc.Miller{H: 1, K: 1}),
				},
				pfc.Peak{
					PlaneDensity: 1.0,
					Location:     wavenumber2,
					Width:        0.01,
					NumPlanes:    pfc.NumEquivalent2D(pfc.Miller{H: 2, K: 0}),
				},
			},
		},
		Field:     "density",
		Prefactor: 1.0,
	}
	elasticProperties := make(map[string]float64)
	elasticProperties["area"] = 68.0

	widthResult := modeSolver.PeakWidths(elasticProperties, []float64{A11, A20})
	term.PairCorrFunc.Peaks[0].Width = widthResult.Widths[0]
	term.PairCorrFunc.Peaks[1].Width = widthResult.Widths[1]

	mixing := IdealMixtureTerm{
		IdealMix:  modeSolver.IdealMix,
		Prefactor: 1.0,
		Field:     "density",
	}

	ft := NewFFTW([]int{N, N})
	excessEnergy := term.GetEnergy(bricks, ft, []int{N, N}) / float64(N*N)
	expectedExcess := -0.5 * float64(term.PairCorrFunc.Peaks[0].NumPlanes) * A11 * A11
	expectedExcess -= 0.5 * float64(term.PairCorrFunc.Peaks[1].NumPlanes) * A20 * A20

	tol := 1e-10
	if math.Abs(excessEnergy-expectedExcess) > tol {
		t.Errorf("Expected excess energy: %f. Got %f\n", expectedExcess, excessEnergy)
	}

	idealEnergy := mixing.GetEnergy(bricks, N*N) / float64(N*N)
	energy := excessEnergy + idealEnergy
	modeSolverEnergy := modeResult.Energy
	if math.Abs(energy-modeSolverEnergy) > tol {
		t.Errorf("Mode solver gives energy %f.\nDirect integral evaluation gives %f\n", modeSolverEnergy, energy)
	}

	// Compress the system and confirm that the energy increases by the expected amount
	delta := 0.001
	term.PairCorrFunc.Peaks[0].Location /= (1.0 - delta)
	term.PairCorrFunc.Peaks[1].Location /= (1.0 - delta)

	newExcess := term.GetEnergy(bricks, ft, []int{N, N}) / float64(N*N)
	dE := newExcess - excessEnergy
	expect := 0.5 * widthResult.Fit["area"] * math.Pow(2.0*delta, 2)
	rtol := 0.01 // One percent tolerance
	if math.Abs(expect-dE) > rtol*expect {
		t.Errorf("Expected increase: %f. Got %f\n", expect, dE)
	}
}

func TestIdealMixTermWithModel(t *testing.T) {
	N := 16
	field := NewField("density", N*N, nil)

	for i := range field.Data {
		field.Data[i] = complex(0.5, 0.0)
	}

	term := IdealMixtureTerm{
		IdealMix:  pfc.IdealMix{C3: 1.0, C4: 1.0},
		Field:     "density",
		Prefactor: 1.0,
		Laplacian: false,
	}

	for _, lap := range []bool{false, true} {
		term.Laplacian = lap

		model := NewModel()
		model.AddField(field)

		// Case 1: Do not register the derived field
		model.RegisterMixedTerm("IDEAL_MIX", &term, nil)
		model.AddEquation("ddensity/dt = IDEAL_MIX")
		model.Init()
		freq := func(i int) []float64 {
			return []float64{0.0, 0.6}
		}

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Should have panicked bacuse of missing field!")
				}
			}()
			model.GetRHS(0, freq, 0.0)
		}()

		// Case 2: Register the missing field
		model.RegisterDerivedField(term.DerivedField(N*N, model.Bricks))
		model.Init()

		// Evaluate the implicit part
		tol := 1e-10
		denum := model.GetDenum(0, freq, 0.0)
		for i := range denum {
			re := real(denum[i])
			im := imag(denum[i])
			expect := 1.0

			if lap {
				expect *= -4.0 * math.Pi * math.Pi * 0.6 * 0.6
			}
			if math.Abs(re-expect) > tol || math.Abs(im) > tol {
				t.Errorf("Expected %f got %f\n", re, expect)
			}
		}

		rhs := model.GetRHS(0, freq, 0.0)
		expect := -0.5*math.Pow(0.5, 2.0) + math.Pow(0.5, 3.0)/3.0
		if lap {
			expect *= -4.0 * math.Pi * math.Pi * 0.6 * 0.6
		}
		for i := range rhs {
			re := real(rhs[i])
			im := imag(rhs[i])
			if math.Abs(re-expect) > tol || math.Abs(im) > tol {
				t.Errorf("Expected %f. Got %f\n", expect, re)
			}
		}
	}

}
