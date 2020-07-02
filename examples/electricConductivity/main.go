package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"math"
	"os"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfutil"
	_ "github.com/mattn/go-sqlite3"
	"gonum.org/v1/gonum/mat"
)

// N is the number of nodes in each direction
const N = 16

// GrainConductivity is a type for representing the conductivity in grains of
// different orientation
type GrainConductivity struct {
	Grain           []pf.Field
	RefConductivity *mat.Dense
	Rotations       []*mat.Dense
}

// RotateConductivity returns the RefConductivity after rotation
func (gc *GrainConductivity) RotateConductivity(rot *mat.Dense) []float64 {
	res := mat.NewDense(2, 2, nil)
	res.Product(rot, gc.RefConductivity, rot.T())
	return []float64{res.At(0, 0), res.At(1, 1), res.At(0, 1)}
}

// Conductivity returns the conductivity at node i
func (gc *GrainConductivity) Conductivity(i int) []float64 {
	sigma := make([]float64, 3)
	weight := 0.0
	for grain := 0; grain < len(gc.Grain); grain++ {
		rotatedSigma := gc.RotateConductivity(gc.Rotations[grain])
		for j := 0; j < 3; j++ {
			sigma[j] += real(gc.Grain[grain].Get(i)) * rotatedSigma[j]
		}
		weight += real(gc.Grain[grain].Get(i))
	}
	for j := 0; j < 3; j++ {
		sigma[j] /= weight
	}
	return sigma
}

func rotMatrix(angle float64) *mat.Dense {
	matrix := mat.NewDense(2, 2, nil)
	matrix.Set(0, 0, math.Cos(angle))
	matrix.Set(0, 1, math.Sin(angle))
	matrix.Set(1, 0, -math.Sin(angle))
	matrix.Set(1, 1, math.Cos(angle))
	return matrix
}

// GetFields return a list of fields that corresponds to the
// different grains
func GetFields(voronoi []int) []pf.Field {
	numGrains := 0
	for i := range voronoi {
		if voronoi[i] > numGrains {
			numGrains = voronoi[i]
		}
	}

	fields := []pf.Field{}
	work := make([]float64, N*N)
	for grain := 0; grain < numGrains; grain++ {
		field := pf.NewField(fmt.Sprintf("Grain%d", grain), N*N, nil)
		for i := range voronoi {
			if voronoi[i] == grain {
				work[i] = 1.0
			}
		}
		pfutil.Blur(&pfutil.RealSlice{Data: work}, []int{N, N}, &pfutil.BoxKernel{Width: 8})
		for i := range work {
			field.Data[i] = complex(work[i], 0.)
		}
		fields = append(fields, field)
	}
	return fields
}

func main() {
	dt := 0.1 // Timestep
	dbName := "grainBoundaryCurrent.db"

	// Define the grains via a Voronoi diagram
	voronoiPts := [][]int{
		{10, 10},
		{5, 50},
		{70, 40},
		{40, 70},
		{20, 100},
		{100, 20},
		{50, 30},
	}
	numGrains := len(voronoiPts)

	domainSize := []int{N, N}
	voronoiNodes := make([]int, len(voronoiPts))
	for i, p := range voronoiPts {
		voronoiNodes[i] = pfutil.NodeIdx(domainSize, p)
	}

	grains := make([]int, N*N)
	pfutil.Voronoi(voronoiNodes, grains, domainSize)

	// Initialize the model and the required fields
	model := pf.NewModel()
	density := pf.NewField("density", N*N, nil)

	// Set up the conductivity for the different grains
	grainCond := GrainConductivity{
		Grain:           GetFields(grains),
		RefConductivity: mat.NewDense(2, 2, []float64{1.0, 0.0, 0.0, 3.0}),
		Rotations:       make([]*mat.Dense, numGrains),
	}

	for i := 0; i < numGrains; i++ {
		angle := float64(i) * math.Pi / 9.0
		grainCond.Rotations[i] = rotMatrix(angle)
	}

	fft := pf.NewFFTW(domainSize)
	charge := pf.ChargeTransport{
		Conductivity:  grainCond.Conductivity,
		ExternalField: []float64{1.0, 0.0},
		Field:         "density",
		FT:            fft,
	}

	model.AddField(density)
	model.RegisterExplicitTerm("MINUS_DIV_CURRENT", &charge, nil)
	model.AddEquation("ddensity/dt = MINUS_DIV_CURRENT")

	solver := pf.NewSolver(&model, domainSize, dt)

	sqlDB, _ := sql.Open("sqlite3", dbName)
	db := pf.FieldDB{
		DB:         sqlDB,
		DomainSize: domainSize,
	}
	solver.AddCallback(db.SaveFields)
	model.Summarize()
	solver.Solve(10, 10)

	// Extract the current
	current := charge.Current(density, len(density.Data), true)

	// Store the current in a csv file
	out, _ := os.Create("current.csv")
	defer out.Close()

	writer := csv.NewWriter(out)
	defer writer.Flush()
	writer.Write([]string{"CurrentX", "CurrentY"})
	for i := range current[0] {
		row := []string{fmt.Sprintf("%f", current[0][i]),
			fmt.Sprintf("%f", current[1][i])}
		writer.Write(row)
	}

}
