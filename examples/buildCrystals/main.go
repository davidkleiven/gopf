// +build ignore

package main

import (
	"fmt"
	"math"

	"github.com/davidkleiven/gopf/pfc"
	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/gonum/mat"
)

// SingleCrystal shows how one can build a single crystal
func SingleCrystal() pfutil.Grid {
	// Initialize the unit cell
	ucell := pfc.UnitCell{
		Cell: pfc.SC2D(32.0),
		Basis: mat.NewDense(2, 2, []float64{0.0, 0.5,
			0.0, 0.5}),
	}

	// Initialize a kernel that represents the field of one atom
	kernel := pfc.GaussianKernel{
		Width: 4.0,
	}

	// Initialize a grid
	grid := pfutil.NewGrid([]int{512, 512})

	// Build the crystal
	pfc.BuildCrystal(ucell, &kernel, &grid)
	return grid
}

func main() {
	grid := SingleCrystal()
	fname := "single_crystal.csv"
	grid.SaveCsv(fname)
	fmt.Printf("Grid written to %s\n", fname)

	// Convert into a grain boundary
	angle := 15.0 * math.Pi / 180.0
	pfc.TiltGB(&grid, angle)
	fname = "grainBoundary.csv"
	grid.SaveCsv(fname)
	fmt.Printf("Grain boundary written to %s\n", fname)
}
