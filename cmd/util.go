package cmd

import (
	"database/sql"
	"fmt"
)

// SimIDWidth is the width of the simulation ID field
const SimIDWidth = 10

// ColWidth is the width of the remaining columns in a table
const ColWidth = 10

// MaxCols is the maximum number number of columns per line
const MaxCols = 10

// CommentWidth is the width of the comments field
const CommentWidth = 96

// singleLine produce a line with "-"
func singleLine(length int) string {
	singleLine := ""
	for i := 0; i < length; i++ {
		singleLine += "-"
	}
	return singleLine
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

// closestWhiteSpace returns the position of the first white space
// to the left of target
func closestWhiteSpace(line string, target int) int {
	for i := target; i >= 0; i-- {
		if line[i] == ' ' {
			return i
		}
	}
	return 0
}
