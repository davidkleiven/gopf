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
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// ValueWidth represents the width of the value field
const ValueWidth = 60

// DescWidth represents the width of the description field
const DescWidth = 16

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show general information/summary of the database",
	Long: `info prints general information about the database.
It prints the number of calcualtions, dimensions of the underlying
simulation domain, the name of all the fields in the database,
the name of all attributes that are attached to the simulation, and
the name of all timeseries that are tracked.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatalf("A database must be given.")
			return
		}

		db, err := sql.Open("sqlite3", args[0])
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}

		// Extract the number of calculations
		rows, err := db.Query("SELECT COUNT (*) FROM simIDs")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		var numCalc int
		for rows.Next() {
			rows.Scan(&numCalc)
		}

		// Extract the names of the fields
		rows, err = db.Query("SELECT DISTINCT name FROM fields")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		fields := []string{}
		var fieldName string
		for rows.Next() {
			rows.Scan(&fieldName)
			fields = append(fields, fieldName)
		}

		// Extract the dimension of the data
		rows, err = db.Query("SELECT MAX(X) FROM positions")
		var nx, ny, nz int
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		for rows.Next() {
			rows.Scan(&nx)
		}

		rows, err = db.Query("SELECT MAX(Y) FROM positions")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		for rows.Next() {
			rows.Scan(&ny)
		}

		rows, err = db.Query("SELECT MAX(Z) FROM positions")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		for rows.Next() {
			rows.Scan(&nz)
		}

		// Extract attributes
		rows, err = db.Query("SELECT DISTINCT key FROM simAttributes")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		attrs := []string{}
		var attr string
		for rows.Next() {
			rows.Scan(&attr)
			attrs = append(attrs, attr)
		}

		rows, err = db.Query("SELECT DISTINCT key FROM simTextAttributes")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		for rows.Next() {
			rows.Scan(&attr)
			attrs = append(attrs, attr)
		}

		// Extract the timeseries
		rows, err = db.Query("SELECT DISTINCT key FROM timeseries")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}

		var tname string
		tnames := []string{}
		for rows.Next() {
			rows.Scan(&tname)
			tnames = append(tnames, tname)
		}

		// Print the summary
		fmt.Printf("%s\n", horizontalLine())
		fmt.Printf("| %-*s | %-*s |\n", DescWidth, "Desc.", ValueWidth, "Value")
		fmt.Printf("%s\n", horizontalLine())
		fmt.Printf("| %-*s | %*d |\n", DescWidth, "Num calc.", ValueWidth, numCalc)
		fmt.Printf("| %-*s | %*s |\n", DescWidth, "Dims", ValueWidth, dimString(nx+1, ny+1, nz+1))
		printList("Fields", fields)
		printList("Attributes", attrs)
		printList("Timeseries", tnames)
		fmt.Printf("%s\n", horizontalLine())
	},
}

func horizontalLine() string {
	var line string
	for i := 0; i < DescWidth+ValueWidth+7; i++ {
		line += "-"
	}
	return line
}

func dimString(nx, ny, nz int) string {
	return fmt.Sprintf("(%d, %d, %d)", nx, ny, nz)
}

func printList(name string, array []string) {
	line := strings.Join(array, ", ")
	desc := name
	for i := 0; i < len(line)/ValueWidth+1; i++ {
		if i > 0 {
			desc = ""
		}
		var str string
		if (i+1)*ValueWidth > len(line) {
			str = line[i*ValueWidth:]
		} else {
			str = line[i*ValueWidth : (i+1)*ValueWidth]
		}
		fmt.Printf("| %-*s | %*s |\n", DescWidth, desc, ValueWidth, str)
	}

}

func init() {
	dbCmd.AddCommand(infoCmd)
}
