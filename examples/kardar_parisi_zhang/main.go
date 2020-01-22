package main

import (
	"flag"
	"math"
	"math/rand"
	"time"

	"github.com/davidkleiven/gopf/pf"
)

func main() {
	prefix := flag.String("prefix", "kpz", "prefix where files should be stored")
	flag.Parse()
	now := time.Now().UnixNano()
	rand.Seed(now)
	lamb := 10.0
	dt := 0.0001
	strength := 0.01 / math.Sqrt(dt)

	N := 128

	model := pf.NewModel()
	height := pf.NewField("height", N*N, nil)
	model.AddField(height)
	noise := pf.WhiteNoise{
		Strength: strength,
	}

	gradSq := pf.NewSquareGradient("height", []int{N, N})
	gradSq.Factor = lamb

	model.RegisterFunction("WHITE_NOISE", noise.Generate)
	model.RegisterUserDefinedTerm("GRAD_SQ", &gradSq, nil)

	model.AddEquation("dheight/dt = LAP height + GRAD_SQ + WHITE_NOISE")

	solver := pf.NewSolver(&model, []int{N, N}, dt)
	out := pf.NewFloat64IO(*prefix)
	solver.AddCallback(out.SaveFields)

	nepoch := 3
	solver.Solve(nepoch, 20)

	pf.WriteXDMF(*prefix+".xdmf", []string{"height"}, "kpz", nepoch, []int{N, N})
}
