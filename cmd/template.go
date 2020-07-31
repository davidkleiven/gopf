package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/davidkleiven/gopf/elasticity"
	"github.com/spf13/cobra"
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Generates a template for the JSON file used as input to the energy calculator",
	Long:  `Generates a template for the input file required for the energy calculation`,
	Run: func(cmd *cobra.Command, args []string) {
		out, err := cmd.Flags().GetString("out")
		if err != nil {
			log.Fatalf("%s\n", err)
			return
		}
		bulkMod := 60.0
		poisson := 0.3
		shear := elasticity.Shear(bulkMod, poisson)

		c11 := bulkMod + 4.0*shear/3.0
		c12 := bulkMod - 2.0*shear/3.0
		f := 2.0
		params := elasticity.StrainEnergyInput{
			HalfA:  10.0,
			HalfB:  10.0,
			HalfC:  10.0,
			Misfit: []float64{0.01, 0.01, 0.01, 0.0, 0.0, 0.0},
			MatPropMatrix: []float64{c11, c12, c12, 0.0, 0.0, 0.0,
				c12, c11, c12, 0.0, 0.0, 0.0,
				c12, c12, c11, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, shear, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, shear, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, shear},
			MatPropInc: []float64{f * c11, f * c12, f * c12, 0.0, 0.0, 0.0,
				f * c12, f * c11, f * c12, 0.0, 0.0, 0.0,
				f * c12, f * c12, f * c11, 0.0, 0.0, 0.0,
				0.0, 0.0, 0.0, f * shear, 0.0, 0.0,
				0.0, 0.0, 0.0, 0.0, f * shear, 0.0,
				0.0, 0.0, 0.0, 0.0, 0.0, f * shear},
			DomainSize:        64,
			ApplyPerturbation: true,
		}

		file, err := json.MarshalIndent(params, "", " ")
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(out, file, 0644)
		if err != nil {
			panic(err)
		}
		log.Printf("Template for input file written to %s\n", out)
	},
}

func init() {
	elastCmd.AddCommand(templateCmd)
	templateCmd.Flags().StringP("out", "o", "elasticTemplate.json", "Filename where the result will be written")
}
