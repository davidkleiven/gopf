/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/csv"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

// contourCmd represents the contour command
var contourCmd = &cobra.Command{
	Use:   "contour",
	Short: "Create a contour plot from csv file",
	Long: `This command creates a contour plot from a csv file of the form

X, Y, Z, field1, field2, field3
0, 1, 2, 0.0, 0.1, -0.4,
0, 1, 1, -0.4, 0.2, 0.6
...

the number of fields can be arbitrarily long. However, the field selected for
plotting is the one passed as an argument. If not passed, the first field
will be selected.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fname, err := cmd.Flags().GetString("fname")
		if err != nil {
			log.Fatalf("Could not retrieve file name: %s\n", err)
			return
		}

		column, err := cmd.Flags().GetString("column")
		if err != nil {
			log.Fatalf("Could not retrieve column name: %s\n", err)
			return
		}

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			log.Fatalf("Could not retrieve outfile: %s\n", err)
			return
		}

		header := readHeader(fname)
		idx := getColIndex(header, column)
		column = header[idx]

		rows := readData(fname, column)
		min, max := dataRange(rows)

		data := NewHeatMapData(rows)
		colormap := moreland.Kindlmann()
		rng := max - min
		colormap.SetMin(min - 0.05*rng)
		colormap.SetMax(max + 0.05*rng)

		plt := plot.New()

		heatImg := fillImage(data, colormap)
		n, m := data.Dims()
		pImg := plotter.NewImage(heatImg, 0, 0, float64(n), float64(m))
		plt.Add(pImg)

		barplt := plot.New()

		plt.X.Label.Text = "x position (\u0394 x)"
		plt.Y.Label.Text = "y position (\u0394 x)"

		bar := plotter.ColorBar{
			ColorMap: colormap,
		}
		barplt.Add(&bar)

		img := vgimg.New(4*vg.Inch, 4*vg.Inch)
		dc := draw.New(img)
		top := draw.Crop(dc, 0.325*vg.Inch, -0.325*vg.Inch, 3.35*vg.Inch, 0.0)
		bottom := draw.Crop(dc, 0.325*vg.Inch, -0.325*vg.Inch, 0.0, -0.65*vg.Inch)
		barplt.HideY()
		barplt.Title.Text = column

		barplt.Draw(top)
		plt.Draw(bottom)

		if err != nil {
			log.Fatalf("Could not open output file: %s\n", err)
		}

		format := strings.ToLower(filepath.Ext(out))
		if len(format) != 0 {
			format = format[1:]
		} else {
			log.Fatalf("Could not extract file extension from outfile.\n")
			return
		}

		png := vgimg.PngCanvas{Canvas: img}

		file, err := os.Create(out)
		if err != nil {
			log.Fatalf("Could not open file: %s\n", err)
		}
		defer file.Close()

		_, err = png.WriteTo(file)
		if err != nil {
			log.Fatalf("Could not write to file: %s\n", err)
		}
		log.Printf("Image written to %s\n", out)
	},
}

func init() {
	rootCmd.AddCommand(contourCmd)

	contourCmd.Flags().StringP("fname", "f", "", "CSV file with the data")
	contourCmd.Flags().StringP("column", "c", "", "Name of the of the column to be plotted. Must be one of the names in the header of the file.")
	contourCmd.Flags().StringP("out", "o", "gopfPlot.png", "Outfile where the resulting image is stored.")
}

// DataRow represents one row
type DataRow struct {
	X, Y, Z int
	Value   float64
}

// HeatMapData implements the XYZ interface
type HeatMapData struct {
	rows   []DataRow
	index  []int
	Nx, Ny int
}

// NewHeatMapData returns a new correctly initialized HeatMapData
func NewHeatMapData(rows []DataRow) HeatMapData {
	heatMap := HeatMapData{
		rows:  rows,
		index: make([]int, len(rows)),
	}

	Nx := 0
	Ny := 0
	for _, row := range rows {
		if row.X > Nx {
			Nx = row.X
		}
		if row.Y > Ny {
			Ny = row.Y
		}
	}
	Nx++
	Ny++
	heatMap.Nx = Nx
	heatMap.Ny = Ny

	for i, row := range rows {
		heatMap.index[index(row.X, row.Y, heatMap.Nx)] = i
	}
	return heatMap
}

func index(x int, y int, Nx int) int {
	return y*Nx + x
}

// Dims returns the dimansion of the data
func (h HeatMapData) Dims() (c, r int) {
	return h.Nx, h.Ny
}

// X returns the x coordinate
func (h HeatMapData) X(c int) float64 {
	return float64(c)
}

// Y returns they coordinate
func (h HeatMapData) Y(r int) float64 {
	return float64(r)
}

// Z returns the value of the field
func (h HeatMapData) Z(c, r int) float64 {
	return h.rows[index(c, r, h.Nx)].Value
}

func getColIndex(header []string, column string) int {
	if column == "" {
		return 3
	}
	for i, v := range header {
		if v == column {
			return i
		}
	}

	log.Printf("No column matched the specified names. Using first field\n")
	return 3
}

func readData(fname string, column string) []DataRow {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Could not read header: %s\n", err)
	}

	if len(header) < 4 {
		log.Fatalf("The length of the header must be at least 4. Read: %v\n", header)
	}

	idx := getColIndex(header, column)
	rows := []DataRow{}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error during read: %s\n", err)
		}

		x, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatalf("Could not convert string to float %s\n", err)
		}

		y, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatalf("Could not convert string to float: %s\n", err)
		}

		z, err := strconv.Atoi(record[2])
		if err != nil {
			log.Fatalf("Could not convert string to float: %s\n", err)
		}

		value, err := strconv.ParseFloat(record[idx], 64)
		if err != nil {
			log.Fatalf("Could not convert string to float: %s\n", err)
		}

		rows = append(rows, DataRow{
			X:     x,
			Y:     y,
			Z:     z,
			Value: value,
		})
	}
	return rows
}

func dataRange(records []DataRow) (float64, float64) {
	minval := records[0].Value
	maxval := records[0].Value
	for _, row := range records {
		if row.Value < minval {
			minval = row.Value
		}

		if row.Value > maxval {
			maxval = row.Value
		}
	}
	return minval, maxval
}

func readHeader(fname string) []string {
	file, err := os.Open(fname)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Could not read header: %s\n", err)
	}
	return header
}

func fillImage(data HeatMapData, cmap palette.ColorMap) *image.RGBA64 {
	n, m := data.Dims()
	img := image.NewRGBA64(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: n, Y: m},
	})

	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			color, err := cmap.At(data.Z(i, j))
			if err != nil {
				log.Fatalf("%s\n", err)
				return img
			}
			img.Set(i, j, color)
		}
	}
	return img
}
