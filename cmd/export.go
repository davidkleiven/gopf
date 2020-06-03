package cmd

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to csv files",
	Long: `This command exports data from the database to simple comma separated files.

Example:

gopf db export mydatabase.db --type timeseries

exports all temperature data to a csv file where the timestep is stored in the first
column and corresponding values in the remaining columns. The first row is a header
that describes the content of the file.

gopf db export mystabaset.db --type fielddata --timestep 0

exports the field data for the zeroth timestep. The three first columns contains the
X, Y and Z position. The remaining columns are field values. The first row is a header
that describes the content of each column.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("A database name must be given.")
			return
		}

		db, err := sql.Open("sqlite3", args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		ftype, err := cmd.Flags().GetString("type")
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		outfile, err := cmd.Flags().GetString("out")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		simid, err := cmd.Flags().GetInt("simid")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		if simid < 0 {
			simid = newestSimulationID(db)
		}

		if outfile == "" {
			outfile = fmt.Sprintf("%s_%d.csv", ftype, simid)
		}

		timestep, err := cmd.Flags().GetInt("timestep")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		switch ftype {
		case "timeseries", "ts":
			exportTimeseries(db, outfile, simid)
		case "field", "fieldData", "fd":
			exportFieldData(db, timestep, simid, outfile)
		default:
			fmt.Printf("Unknown export type %s\n", ftype)
		}
	},
}

func init() {
	dbCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("type", "t", "timeseries", "If timeries or field data should be exported.")
	exportCmd.Flags().StringP("out", "o", "", "Name of the output file. If empty, a name will be crafted from the other arguments.")
	exportCmd.Flags().IntP("simid", "i", -1, "Simulation ID. If negative, the newest ID will be used.")
	exportCmd.Flags().IntP("timestep", "s", 0, "Timestep to export (only relevant if type is field).")
}

func exportTimeseries(db *sql.DB, outfile string, simid int) {
	// Extract keys
	rows, err := db.Query("SELECT DISTINCT key FROM timeseries WHERE simId=?", simid)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	keys := []string{}
	var key string
	for rows.Next() {
		rows.Scan(&key)
		keys = append(keys, key)
	}

	// Extract maximum number timestep
	rows, err = db.Query("SELECT MAX(timestep) FROM timeseries WHERE simId=?", simid)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	var maxStep int
	for rows.Next() {
		rows.Scan(&maxStep)
	}

	dataArray := make([]map[string]float64, maxStep+1)
	for i := range dataArray {
		dataArray[i] = make(map[string]float64)
	}

	rows, err = db.Query("SELECT key,value,timestep FROM timeseries WHERE simId=?", simid)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	var timestep int
	var value float64

	for rows.Next() {
		rows.Scan(&key, &value, &timestep)
		dataArray[timestep][key] = value
	}

	out, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	defer out.Close()
	writer := csv.NewWriter(out)
	defer writer.Flush()

	writer.Write(keys)
	record := make([]string, len(keys))
	for _, item := range dataArray {
		for j, k := range keys {
			record[j] = fmt.Sprintf("%f", item[k])
		}
		writer.Write(record)
	}
	fmt.Printf("Timeseries written to %s\n", outfile)
}

func exportFieldData(db *sql.DB, ts int, simid int, outfile string) {
	rows, err := db.Query("SELECT DISTINCT name FROM fields WHERE simId=? AND timestep=?", simid, ts)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	keys := []string{}
	var key string
	for rows.Next() {
		rows.Scan(&key)
		keys = append(keys, key)
	}

	positions := [][]int{}
	rows, err = db.Query("SELECT X,Y,Z FROM positions ORDER BY id")
	var x, y, z int
	for rows.Next() {
		rows.Scan(&x, &y, &z)
		positions = append(positions, []int{x, y, z})
	}

	data := make([]map[string]float64, len(positions))
	for i := range data {
		data[i] = make(map[string]float64)
	}

	rows, err = db.Query("SELECT name,value,positionId FROM fields WHERE simId=? AND timestep=?", simid, ts)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	var name string
	var value float64
	var positionID int
	for rows.Next() {
		rows.Scan(&name, &value, &positionID)
		data[positionID][name] = value
	}

	out, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	defer out.Close()
	writer := csv.NewWriter(out)
	defer writer.Flush()

	header := make([]string, 3+len(keys))
	header[0] = "X"
	header[1] = "Y"
	header[2] = "Z"
	copy(header[3:], keys)
	writer.Write(header)
	record := make([]string, len(keys)+3)
	for i, item := range data {
		for j, v := range positions[i] {
			record[j] = fmt.Sprintf("%d", v)
		}
		for j, k := range keys {
			record[3+j] = fmt.Sprintf("%f", item[k])
		}
		writer.Write(record)
	}
	fmt.Printf("Field data written to %s\n", outfile)
}