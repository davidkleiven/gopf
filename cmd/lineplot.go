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
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// lineplotCmd represents the lineplot command
var lineplotCmd = &cobra.Command{
	Use:   "lineplot",
	Short: "Command for plotting fields a long a line",
	Long: `Lineplot plots fields along a line.

Example:

gopf lineplot -f data.csv -d field1,field2,field3 -o lineplot.png -y 64 -z 0

will plot the field names field1, field2 and field3 from the datafile, along the
line defined by y = 64 and z = 0. If not fields are given, all fields in the file
will be plotted.

The CSV file must be formatted as

X, Y, Z, field1, field2, field3
0, 1, 0, 0.0, 1.0, 2.0
1, 0, 0, 0.3, -0.2, 1.0
...

this the three first columns gives the position (x, y, z) position of the point, and
successive columns holds the value of the field indicated in the header.

To plot data along a line, at least two of -x, -y and -z must be specified.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fname, err := cmd.Flags().GetString("fname")
		if err != nil {
			log.Fatalf("Error when loading filename: %s\n", err)
			return
		}

		if fname == "" {
			log.Fatalf("No filename given\n")
			return
		}

		out, err := cmd.Flags().GetString("out")
		if err != nil {
			log.Fatalf("Error when parsing outfile: %s\n", err)
			return
		}

		x, err := cmd.Flags().GetInt("x")
		if err != nil {
			log.Fatalf("Could not read x: %s\n", err)
			return
		}

		y, err := cmd.Flags().GetInt("y")
		if err != nil {
			log.Fatalf("Could not read y: %s\n", err)
			return
		}

		z, err := cmd.Flags().GetInt("z")
		if err != nil {
			log.Fatalf("Could not read z: %s\n", err)
			return
		}

		if numSet(x, y, z) != 2 {
			log.Fatalf("Exactly two of x, y and set must be specified.\n")
			return
		}

		fields, err := cmd.Flags().GetString("fields")

		var fieldArray []string
		if fields == "" {
			fieldArray = readHeader(fname)[3:]
		} else {
			splitted := strings.Split(fields, ",")
			for i := 4; i < len(splitted); i++ {
				fieldArray = append(fieldArray, splitted[i])
			}
		}

		plt, err := plot.New()
		if err != nil {
			log.Fatalf("Error when creatting plot: %s\n", err)
			return
		}

		for _, name := range fieldArray {
			rows := readData(fname, name)
			data := lineData(rows, x, y, z)
			fmt.Printf("NUM DATA: %d\n", len(data))
			line, err := plotter.NewLine(data)
			if err != nil {
				log.Fatalf("Could not create line: %s\n", err)
				continue
			}
			plt.Add(line)
			plt.Legend.Add(name, line)
		}

		plt.X.Label.Text = "Position (\u0394 x)"
		plt.Y.Label.Text = "Field value"

		err = plt.Save(4*vg.Inch, 3*vg.Inch, out)
		if err != nil {
			log.Fatalf("Error when saving: %s\n", err)
		} else {
			log.Printf("Plot written to %s\n", out)
		}
	},
}

func numSet(x, y, z int) int {
	numSet := 0
	if x != -1 {
		numSet++
	}

	if y != -1 {
		numSet++
	}

	if z != -1 {
		numSet++
	}
	return numSet
}

func init() {
	rootCmd.AddCommand(lineplotCmd)

	lineplotCmd.Flags().StringP("fname", "f", "", "CSV file to read data from")
	lineplotCmd.Flags().StringP("data", "d", "", "Comma separated list of fields that should be plotted")
	lineplotCmd.Flags().StringP("out", "o", "lineplot.png", "Outfile (default lineplot.png)")
	lineplotCmd.Flags().IntP("x", "x", -1, "X-position of the target line")
	lineplotCmd.Flags().IntP("y", "y", -1, "Y-position of the target line")
	lineplotCmd.Flags().IntP("z", "z", -1, "Z-position of the target line")
}

func includeRow(row DataRow, x int, y int, z int) bool {
	include := true
	if x != -1 {
		include = include && row.X == x
	}

	if y != -1 {
		include = include && row.Y == y
	}

	if z != -1 {
		include = include && row.Z == z
	}
	return include
}

// absicca returns 0, 1 or 2 if the data should be plotted along the x, y or z axis, respectively
func absissaIndex(x int, y int, z int) int {
	if x == -1 {
		return 0
	}

	if y == -1 {
		return 1
	}

	return 2
}

func lineData(rows []DataRow, x int, y int, z int) plotter.XYs {
	N := len(rows)
	xys := plotter.XYs{}
	for i := N - 1; i >= 0; i-- {
		if includeRow(rows[i], x, y, z) {
			posArray := []float64{float64(rows[i].X), float64(rows[i].Y), float64(rows[i].Z)}
			xys = append(xys, plotter.XY{
				X: posArray[absissaIndex(x, y, z)],
				Y: rows[i].Value,
			})
		}
	}
	return xys
}
