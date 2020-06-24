package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists content of the database",
	Long: `This command can be used to list contents of the database.

Examples:

gopf db list mydatabase.db -c simid

lists the simulation IDs along with the time of creation. 

gopf db list mydatabase.db -c comment

lists all the comments along with the simulations ID.

In all cases, the items are sorted according to the time of creation
such that the newest entries appear first.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("Database name must be specified\n")
			return
		}

		content, err := cmd.Flags().GetString("content")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		max, err := cmd.Flags().GetInt("max")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		sqlDB, err := sql.Open("sqlite3", args[0])

		switch content {
		case "simId", "simid", "id":
			showSimulationIds(sqlDB, max)
		case "comment":
			showComments(sqlDB, max)
		default:
			fmt.Printf("Unknown option %s\n", content)
		}
	},
}

func init() {
	dbCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("content", "c", "simId", "Specify what should be listed. Can be one of simId or comment.")
	listCmd.Flags().IntP("max", "m", 20, "Maximum number of items that are listed")
}

func showSimulationIds(db *sql.DB, max int) {
	rows, err := db.Query("SELECT COUNT(*) FROM simIds")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	var numRows int
	for rows.Next() {
		rows.Scan(&numRows)
	}
	if max < numRows {
		fmt.Printf("Showing %d of %d rows\n", max, numRows)
	}

	sql := "SELECT simID,creationTime FROM simIds ORDER BY creationTime DESC LIMIT "
	sql += fmt.Sprintf("%d", max)
	rows, err = db.Query(sql)

	fmt.Printf("-----------------------------------------------------------\n")
	fmt.Printf("| Sim Id                  | Creation time                 |\n")
	fmt.Printf("-----------------------------------------------------------\n")
	var simID int
	var timestamp string
	for rows.Next() {
		rows.Scan(&simID, &timestamp)
		fmt.Printf("| %-23d | %-29s |\n", simID, timestamp)
	}
	fmt.Printf("-----------------------------------------------------------\n")
}

func showComments(db *sql.DB, max int) {
	rows, err := db.Query("SELECT COUNT(*) FROM simIds")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	var numRows int
	for rows.Next() {
		rows.Scan(&numRows)
	}
	if max < numRows {
		fmt.Printf("Showing %d of %d rows\n", max, numRows)
	}

	simIds := []int{}
	sql := "SELECT simID,creationTime FROM simIds ORDER BY creationTime DESC LIMIT "
	sql += fmt.Sprintf("%d", max)
	rows, err = db.Query(sql)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	var simID int
	var creationTime string

	for rows.Next() {
		rows.Scan(&simID, &creationTime)
		simIds = append(simIds, simID)
	}

	sql = "SELECT value from comments WHERE simID=?"
	var comment string
	for _, v := range simIds {
		rows, err = db.Query(sql, v)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		for rows.Next() {
			rows.Scan(&comment)
		}

		fmt.Printf("Sim Id %d:\n%s\n\n", v, comment)
	}
}
