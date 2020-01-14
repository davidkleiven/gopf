// CLI tool that writes a template for the input file required for ellipsoidal
// elasticity calculations
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/davidkleiven/gopf/elasticity"
)

func main() {
	outfile := flag.String("out", "elasticInputfile.json", "Filename where for the template")
	flag.Parse()

	bulkMod := 60.0
	poisson := 0.3
	shear := elasticity.Shear(bulkMod, poisson)

	c11 := bulkMod + 4.0*shear/3.0
	c12 := bulkMod - 2.0*shear/3.0

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
		DomainSize: 64,
	}

	file, err := json.MarshalIndent(params, "", " ")
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(*outfile, file, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Template for input file written to %s\n", *outfile)
}
