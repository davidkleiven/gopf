package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/davidkleiven/gopf/elasticity"
	"github.com/spf13/cobra"
)

// energyCmd represents the energy command
var energyCmd = &cobra.Command{
	Use:   "energy",
	Short: "Calculates the elastic energy of ellipsoidal inclusions.",
	Long: `Calculates the elastic energy of ellipsoidal inclusions. Physical properties is 
passed via a JSON file. See gopf elast templatate for further information.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		infile, err := cmd.Flags().GetString("input")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		file, err := os.Open(infile)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		byteValue, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}

		var params elasticity.StrainEnergyInput
		json.Unmarshal(byteValue, &params)
		energy := elasticity.CalculateStrainEnergy(params)
		log.Printf("Elastic energy per volume inclusion: %f\n", energy)
	},
}

func init() {
	elastCmd.AddCommand(energyCmd)
	energyCmd.Flags().StringP("input", "i", "", "JSON file containing input data")
}
