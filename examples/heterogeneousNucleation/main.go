// +build ignore

package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfutil"
)

func theta(x float64, y float64) float64 {
	v := math.Tanh((x + 0.2*math.Cos(4.0*math.Pi*y)) / 0.01)
	gamma := 0.99
	return 1.0 - gamma + gamma*math.Pow(v, 4)
}

func position(i int, N int) (float64, float64) {
	pos := pfutil.Pos([]int{N, N}, i)
	x := float64(pos[0])/float64(N) - 0.5
	y := float64(pos[1])/float64(N) - 0.5
	return x, y
}

// SpatialVaryingLap implements the term f(x, y) LAP phi
type SpatialVaryingLap struct {
	Gamma float64
	Ft    *pfutil.FFTWWrapper
}

// Construct builds the fourier transformed right hand side
func (s *SpatialVaryingLap) Construct(b map[string]pf.Brick) pf.Term {
	return func(freq pf.Frequency, t float64, work []complex128) {
		// Phi is now in the fourier domain
		brick := b["phi"]
		for i := range work {
			work[i] = brick.Get(i)
		}
		lap := pf.LaplacianN{Power: 1}
		lap.Eval(s.Ft.Freq, work)

		// Inverse FFT
		s.Ft.IFFT(work)
		pfutil.DivRealScalar(work, float64(len(work)))

		// Multiply by the spatially varying function
		for i := range work {
			x, y := position(i, s.Ft.Dimensions[0])
			work[i] *= complex(s.Gamma*theta(x, y), 0.0)
		}

		// Forward FFT
		s.Ft.FFT(work)
	}
}

// OnStepFinished does nothing, since we don't need to update anything after
// each time steps
func (s *SpatialVaryingLap) OnStepFinished(t float64, b map[string]pf.Brick) {}

func insertSemiCircle(array pfutil.MutableSlice, R float64, center []float64, N int) {
	for i := 0; i < array.Len(); i++ {
		array.Set(i, -1.0)
		x, y := position(i, N)
		rSq := (x-center[0])*(x-center[0]) + (y-center[1])*(y-center[1])
		if rSq < R*R && x >= center[0] {
			array.Set(i, 1.0)
		}
	}
}

func insertTorus(array pfutil.MutableSlice, R1 float64, R2 float64, center []float64, N int) {
	for i := 0; i < array.Len(); i++ {
		x, y := position(i, N)
		rSq := (x-center[0])*(x-center[0]) + (y-center[1])*(y-center[1])
		if rSq >= R1*R1 && rSq <= R2*R2 && x >= center[0] {
			array.Set(i, 1.0)
		}
	}
}

func main() {
	dim := flag.Int("dim", 128, "Size of the domain")
	innerR := flag.Float64("innerR", 4.0, "Inner radius in the torous used to define the initial growth direction")
	outerR := flag.Float64("outerR", 8.0, "Outer radius used to define the initial growth direction")
	folder := flag.String("folder", "", "Folder where output files will be stored. If empty, no files are stored")
	steps := flag.Int("step", 10, "Number of steps per epoch")
	epoch := flag.Int("epoch", 10, "Number of epochs")
	dt := flag.Float64("dt", 0.01, "Timestep")
	flag.Parse()
	N := *dim

	phi := pf.NewField("phi", N*N, nil)
	rho := 0.05
	gamma := 0.5
	ft := pfutil.NewFFTW([]int{N, N})
	spatialLap := SpatialVaryingLap{
		Gamma: gamma,
		Ft:    ft,
	}

	model := pf.NewModel()

	// Initialize the fields. We first start from a circular droplet of radius R
	// The initial growth direction is an outward torus
	R := 0.5 * (*innerR + *outerR)
	insertSemiCircle(&pfutil.RealPartSlice{Data: phi.Data}, R, []float64{-0.2, 0.0}, N)
	orientation := make([]float64, N*N)
	insertTorus(&pfutil.RealSlice{Data: orientation}, *innerR, *outerR, []float64{-0.2, 0.0}, N)

	model.AddField(phi)
	model.RegisterFunction("MINUS_CHEM_POT", func(i int, bricks map[string]pf.Brick) complex128 {
		phi := bricks["phi"].Get(i)
		x, y := position(i, N)

		return (1.0 - phi*phi) * (phi + complex(3.0*rho/4.0, 0.0)) * complex(theta(x, y), 0.0)
	})
	model.RegisterExplicitTerm("SPATIAL_LAP", &spatialLap, nil)

	model.AddEquation("dphi/dt = MINUS_CHEM_POT + SPATIAL_LAP")

	sdd := pf.NewSDD([]int{N, N}, &model)
	sdd.SetInitialOrientation(orientation)
	sdd.TimeConstants.Orientation = 1.0
	sdd.TimeConstants.DimerLength = 1.0
	sdd.InitDimerLength = 1.0
	sdd.MinDimerLength = 5e-6
	sdd.Dt = *dt

	solver := pf.NewSolver(&model, []int{N, N}, *dt)
	solver.Stepper = &sdd

	// Add callbacks to log the progress
	if *folder != "" {
		csvIO := pf.CsvIO{
			Prefix:     *folder + "/phi",
			DomainSize: []int{N, N},
		}
		solver.AddCallback(csvIO.SaveFields)

		solver.AddCallback(func(s *pf.Solver, epoch int) {
			sdd.SaveOrientation(fmt.Sprintf(*folder+"/orientation%d.csv", epoch))
		})

		// Log SDD progress
		sddProgress, err := os.Create(*folder + "/sddProgress.csv")
		if err != nil {
			panic(err)
		}
		sdd.Monitor.LogFile = sddProgress

		solver.AddCallback(sdd.Monitor.Log)
	}

	solver.Solve(*epoch, *steps)
}
