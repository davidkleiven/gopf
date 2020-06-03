package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

// commentCmd represents the comment command
var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Show or update a comment of a given simulation ID",
	Long: `This command lists or updates the comment associated wiht a simulation.

Example:

gopf db comment mydatabase.db --simid 346

show the comment associated with the sumlation with id 346

gopf db comment mydatabase.db --simid 346 --new "Another comment"

updates the comment associated with the simulation with ID 346.

If simid is not given (or is a negative number), the newest simulation
is used.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("A database name must be given.")
			return
		}

		sqlDB, err := sql.Open("sqlite3", args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		simid, err := cmd.Flags().GetInt("simid")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		newMsg, err := cmd.Flags().GetString("new")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		if simid < 0 {
			simid = newestSimulationID(sqlDB)
		}

		if newMsg != "" {
			statement, err := sqlDB.Prepare("UPDATE comments SET value=? WHERE simId=?")
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			statement.Exec(newMsg, simid)
			fmt.Printf("Comment of id %d was successfully updated\n", simid)
		} else {
			rows, err := sqlDB.Query("SELECT value FROM comments WHERE simId=?", simid)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			var comment string
			for rows.Next() {
				rows.Scan(&comment)
			}
			fmt.Printf("Sim id %d:\n%s\n", simid, comment)
		}

	},
}

func init() {
	dbCmd.AddCommand(commentCmd)
	commentCmd.Flags().IntP("simid", "i", -1, "Simulation ID. If negative, the last simulation will be shown.")
	commentCmd.Flags().StringP("new", "n", "", "New comment. If empty, the existing comment will be shown.")
}

// newestSimulationID returns the ID of the newest simulation
func newestSimulationID(db *sql.DB) int {
	rows, err := db.Query("SELECT simId FROM simIds ORDER BY creationTime DESC LIMIT 1")
	if err != nil {
		fmt.Printf("%s\n", err)
		return 0
	}

	var simID int
	for rows.Next() {
		rows.Scan(&simID)
	}
	return simID
}
