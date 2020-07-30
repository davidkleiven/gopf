// +build ignore

package main

import (
	"math"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfutil"
)

func main() {
	N := 128
	grid := pfutil.NewGrid([]int{N, N})

	// Create a 30x30 square
	square := pfutil.Box{Diagonal: []float64{30.0, 30.0}}

	// Translate the square
	trans := pfutil.Translation([]float64{-25.0, -25.0})

	// Draw the square
	pfutil.Draw(&square, &grid, &trans, 1.0)

	// Create a square rotated by 45 degrees
	trans = pfutil.Translation([]float64{-35.0, -70.0})
	trans.Append(pfutil.RotZ(math.Pi / 4.0))
	pfutil.Draw(&square, &grid, &trans, 1.0)

	// Create a rectangle
	trans = pfutil.Translation([]float64{-35.0, -110.0})
	trans.Append(pfutil.Scaling([]float64{1.0, 2.0}))
	pfutil.Draw(&square, &grid, &trans, 1.0)

	// Create rectangle rotated 45 deg
	trans = pfutil.Translation([]float64{-85.0, -25.0})
	trans.Append(pfutil.RotZ(math.Pi / 4.0))
	trans.Append(pfutil.Scaling([]float64{1.0, 2.0}))
	pfutil.Draw(&square, &grid, &trans, 1.0)

	// Create a circle
	circle := pfutil.Ball{Radius: 15.0}
	trans = pfutil.Translation([]float64{-85.0, -65.0})
	pfutil.Draw(&circle, &grid, &trans, 1.0)

	// Create an ellipse
	trans = pfutil.Translation([]float64{-85.0, -110.0})
	trans.Append(pfutil.Scaling([]float64{2.0, 1.0}))
	pfutil.Draw(&circle, &grid, &trans, 1.0)

	// Blur the result
	pfutil.Blur(&pfutil.RealSlice{Data: grid.Data}, grid.Dims, &pfutil.BoxKernel{Width: 3})

	// Save grid to csv
	pf.SaveCsv("shapes.csv", []pf.CsvData{{
		Name: "Indicator",
		Data: &pfutil.RealSlice{Data: grid.Data},
	}}, grid.Dims)
}
