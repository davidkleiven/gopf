package cmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

// SimIDWidth is the width of the simulation ID field
const SimIDWidth = 10

// ColWidth is the width of the remaining columns
const ColWidth = 10

// MaxCols is the maximum number number of columns per line
const MaxCols = 10

func totalWidth() int {
	return SimIDWidth + MaxCols*ColWidth
}

// attrCmd represents the attr command
var attrCmd = &cobra.Command{
	Use:   "attr",
	Short: "Command for listing simulation attributes",
	Long: `This commands list the simulation attributes.

Examples:

gopf db attr mydatabase.db

lists all attributes for all simulation IDs in the database.

gopf db attr mydatabase.db --unique

lists all attribute names that is present in at least one simulation.

gopf db attr mydatabase.db --simid 346

lists all attributes that that belongs to the simulation with ID 346.

gopf db attr mydatabase.db --name concentration

lists the value of the attribute concentration for all simulations where
it is present.

gopf db attr mydatabase.db --simid 346 --name concentration

lists the value of concentration for the simulation with ID 346.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Printf("A database must be submitted")
			return
		}

		db, err := sql.Open("sqlite3", args[0])
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		listUnique, err := cmd.Flags().GetBool("unique")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		if listUnique {
			listUniqueAttributes(db)
			return
		}

		simid, err := cmd.Flags().GetInt("simid")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		if name == "" {
			listAttributes(db, simid)
		} else {
			listSingleAttribute(db, simid, name)
		}
	},
}

func init() {
	dbCmd.AddCommand(attrCmd)
	attrCmd.Flags().IntP("simid", "i", -1, "Simulation ID, if negative all IDs will be listed")
	attrCmd.Flags().StringP("name", "n", "", "Attribute to be shown. If not given, all attributes are listed")
	attrCmd.Flags().BoolP("unique", "u", false, "If given, all unique names are listed")
}

func listUniqueAttributes(db *sql.DB) {
	rows, err := db.Query("SELECT key FROM simAttributes UNION SELECT key FROM simTextAttributes")
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Unique attribute names:\n")
	var key string
	for rows.Next() {
		rows.Scan(&key)
		fmt.Printf("%s\n", key)
	}
}

func singleLine(length int) string {
	singleLine := ""
	for i := 0; i < length; i++ {
		singleLine += "-"
	}
	return singleLine
}

// listAttributes lists the attributes. If simulationID is greater than 0, only attributes
// belonging to the passed ID is printed. If it is negative, all simulation IDs are included.
func listAttributes(db *sql.DB, simulationID int) {
	var rows *sql.Rows
	var err error
	if simulationID > 0 {
		rows, err = db.Query("SELECT key, value, simId FROM simAttributes WHERE simId=?", simulationID)
	} else {
		rows, err = db.Query("SELECT key, value, simId FROM simAttributes")
	}

	type pair struct {
		key   string
		value string
	}

	// Create map where simId is the key, and each item is a list of pairs
	kvp := make(map[int]map[string]string)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	header := make(map[string]bool) // map[string]bool is the idiomatic way of making a set in Go

	// Go through the float attributes
	var key string
	var value float64
	var simID int
	for rows.Next() {
		rows.Scan(&key, &value, &simID)
		if _, ok := kvp[simID]; !ok {
			kvp[simID] = make(map[string]string)
		}
		kvp[simID][key] = fmt.Sprintf("%f", value)
		header[key] = true
	}

	// Go through the text attributes
	var txtVal string

	if simulationID < 0 {
		rows, err = db.Query("SELECT key, value, simId FROM simTextAttributes")
	} else {
		rows, err = db.Query("SELECT key, value, simId FROM simTextAttributes WHERE simId=?", simulationID)
	}

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for rows.Next() {
		rows.Scan(&key, &txtVal, &simID)
		header[key] = true
		if _, ok := kvp[simID]; !ok {
			kvp[simID] = make(map[string]string)
		}
		kvp[simID][key] = txtVal
	}

	headerKeys := set2slice(header)
	for tableIdx := 0; tableIdx <= len(headerKeys)/MaxCols; tableIdx++ {
		var currentHeader []string
		if (tableIdx+1)*MaxCols > len(headerKeys) {
			currentHeader = headerKeys[tableIdx*MaxCols:]
		} else {
			currentHeader = headerKeys[tableIdx*MaxCols : (tableIdx+1)*MaxCols]
		}
		line := fmt.Sprintf("| %-*s |", SimIDWidth, "Sim. ID")
		for _, v := range currentHeader {
			line += fmt.Sprintf(" %-*s |", ColWidth, v)
		}

		fmt.Printf("%s\n", singleLine(len(line)))
		fmt.Printf("%s\n", line)
		fmt.Printf("%s\n", singleLine(len(line)))

		for id, attr := range kvp {
			line := fmt.Sprintf("| %-*d |", SimIDWidth, id)
			for _, key := range currentHeader {
				value, ok := attr[key]
				if !ok {
					value = ""
				}
				line += fmt.Sprintf(" %*s |", ColWidth, value)
			}
			fmt.Printf("%s\n", line)
		}
		fmt.Printf("%s\n", singleLine(len(line)))
	}
}

func listSingleAttribute(db *sql.DB, simulationID int, name string) {
	var rows *sql.Rows
	var err error
	if simulationID > 0 {
		rows, err = db.Query("SELECT value, simId FROM simAttributes WHERE key=? AND simId=?", name, simulationID)
	} else {
		rows, err = db.Query("SELECT value, simId FROM simAttributes WHERE key=?", name)
	}

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	var value float64
	var simID int
	for rows.Next() {
		rows.Scan(&value, &simID)
		fmt.Printf("Sim id %-10d %-20s = %6.6f\n", simID, name, value)
	}

	if simulationID > 0 {
		rows, err = db.Query("SELECT value, simId FROM simTextAttributes WHERE key=? AND simId=?", name, simulationID)
	} else {
		rows, err = db.Query("SELECT value, simId FROM simTextAttributes WHERE key=?", name)
	}

	var txtValue string
	for rows.Next() {
		rows.Scan(&value, &simID)
		fmt.Printf("Sim id %-10d %-20s = %12s\n", simID, name, txtValue)
	}
}

func set2slice(set map[string]bool) []string {
	res := []string{}
	for k := range set {
		res = append(res, k)
	}
	return res
}
