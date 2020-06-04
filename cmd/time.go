package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

// timeCmd represents the time command
var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Prints the creation time of a simulation ID",
	Long: `This command prints the creation time of a given simulation ID.

Example:

gopf db time mydatabase.db -i 456

prints the creation time of the simulation with ID 456.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("A database name must be given")
			return
		}

		db, err := sql.Open("sqlite3", args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		simid, err := cmd.Flags().GetInt("simid")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		if simid < 0 {
			simid = newestSimulationID(db)
		}

		rows, err := db.Query("SELECT creationTime FROM simIds WHERE simId=?", simid)
		var creationTime string
		for rows.Next() {
			rows.Scan(&creationTime)
		}
		fmt.Printf("Simulation %d was created at %s\n", simid, creationTime)
	},
}

func init() {
	dbCmd.AddCommand(timeCmd)

	timeCmd.Flags().IntP("simid", "i", -1, "Simulation ID. If negative, the creation time of the last entry will be shown.")
}
