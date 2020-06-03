package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// dbCmd represents the db command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "List/extract content of a gopf database",
	Long: `
	db is a command for instpecting the contents of a gopf database.
	It can be used to export data to simpler formats for example csv files.
	List content and perform simple queries.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("db called")
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
