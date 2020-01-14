package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/davidkleiven/gopf/elasticity"
)

func main() {
	inputFile := flag.String("input", "", "Input file. See template generated by the gopf-elast-input command")
	flag.Parse()

	file, err := os.Open(*inputFile)
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
	fmt.Printf("Elastic energy per volume inclusion: %f\n", energy)
}
