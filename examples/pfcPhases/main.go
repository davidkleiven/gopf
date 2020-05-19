package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfc"
	"github.com/davidkleiven/gopf/pfutil"
)

// EnergyObserver is a type that calculates the free energy
type EnergyObserver struct {
	Ideal      pf.IdealMixtureTerm
	Excess     pf.PairCorrlationTerm
	DomainSize []int
	LastEnergy float64
}

// Calculate calculates the energy (normalized by volume)
func (eo *EnergyObserver) Calculate(s *pf.Solver, epoch int) {
	N := pfutil.ProdInt(eo.DomainSize)
	idealE := eo.Ideal.GetEnergy(s.Model.Bricks, N)
	excess := eo.Excess.GetEnergy(s.Model.Bricks, s.FT, eo.DomainSize)
	eo.LastEnergy = (idealE + excess) / float64(N)
	fmt.Printf("Last energy %f\n", eo.LastEnergy)
}

func main() {
	effTemp := flag.Float64("temp", 0.0, "Effective temperature")
	meanDensity := flag.Float64("conc", 0.0, "Mean density")
	start := flag.Int("start", 0, "Epoch to start from")
	folderVar := flag.String("folder", "./", "Folder where data should be stored")
	flag.Parse()
	folder := *folderVar
	prefix := fmt.Sprintf("pfc_%d_%d", int(*effTemp)*1000.0, int(*meanDensity*1000.0))
	N := 128
	a := 16.0
	dt := 0.1
	nepoch := 10
	nsteps := 1000

	field := pf.NewField("density", N*N, nil)

	if *start != 0 {
		fname := fmt.Sprintf("data/%s_density_%d.bin", prefix, *start-1)
		data := pf.LoadFloat64(fname)
		for i := range field.Data {
			field.Data[i] = complex(data[i], 0.0)
		}
	} else {
		// Randomize starting point
		for i := range field.Data {
			field.Data[i] = complex(0.3*(2.0*rand.Float64()-1.0)+*meanDensity, 0.0)
		}
	}

	// Two-peak model
	term := pf.PairCorrlationTerm{
		PairCorrFunc: pfc.ReciprocalSpacePairCorrelation{
			EffTemp: *effTemp,
			Peaks:   pfc.SquareLattice2D(0.02, a),
		},
		Field:     "density",
		Prefactor: 1.0,
		Laplacian: true,
	}

	// Mixing energy
	ideal := pf.IdealMixtureTerm{
		IdealMix: pfc.IdealMix{
			C3: 1.0,
			C4: 1.0,
		},
		Field:     "density",
		Prefactor: 1.0,
		Laplacian: true,
	}

	model := pf.NewModel()
	model.AddField(field)
	model.RegisterFunction("IDEAL", ideal.Eval)
	model.RegisterImplicitTerm("EXCESS", &term, nil)
	model.RegisterMixedTerm("IDEAL", &ideal, []pf.DerivedField{ideal.DerivedField(N*N, model.Bricks)})
	model.AddEquation("ddensity/dt = IDEAL + EXCESS")

	solver := pf.NewSolver(&model, []int{N, N}, dt)
	solver.StartEpoch = *start
	model.Summarize()
	writer := pf.NewFloat64IO(folder + prefix)
	solver.AddCallback(writer.SaveFields)

	// Add an energy observer
	observer := EnergyObserver{
		Ideal:      ideal,
		Excess:     term,
		DomainSize: []int{N, N},
	}
	solver.AddCallback(observer.Calculate)
	solver.Solve(nepoch, nsteps)
	xdmfFile := folder + fmt.Sprintf("pfc_%d_%d.xdmf", int(*effTemp*1000.0), int(*meanDensity*1000))
	pf.WriteXDMF(xdmfFile, []string{"density"}, prefix, *start+nepoch, []int{N, N})

	// Append the energy to file
	f, err := os.OpenFile("pfcPhasesEnergies.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not open file!")
	}
	out := fmt.Sprintf("%f,%f,%f\n", *effTemp, *meanDensity, observer.LastEnergy)
	if _, err := f.Write([]byte(out)); err != nil {
		f.Close()
		fmt.Printf("Error when writing!")
	}
	if err := f.Close(); err != nil {
		fmt.Printf("Could not close the logfile!")
	}
}
