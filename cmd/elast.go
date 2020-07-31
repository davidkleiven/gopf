package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// elastCmd represents the elast command
var elastCmd = &cobra.Command{
	Use:   "elast",
	Short: "Command for performing elasticity calculations",
	Long:  `Command for performing elasticity calculations`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("elast called")
	},
}

func init() {
	rootCmd.AddCommand(elastCmd)
}
