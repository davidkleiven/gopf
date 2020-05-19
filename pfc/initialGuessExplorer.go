package pfc

import (
	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/gonum/optimize"
)

// InitialGuessExplorer is a type that is used to generate initial guesses
// for local minimzation algorithms
type InitialGuessExplorer interface {
	// Next return the next initial guess. It should return
	// nil when all initial guess has been explored
	Next() []float64

	// OnMinimzationFinished is a callback method that is called everytime
	// the local minimizer finds a solution
	OnMinimizationFinished(res *optimize.Result)
}

// HyperOctantExplorer tries initial conditions starting from all
// combinations of +- X0
type HyperOctantExplorer struct {
	// Starting positions
	X0 []float64

	// Product
	prod pfutil.Product
}

// Next returns the next hyper parameter
func (hoe *HyperOctantExplorer) Next() []float64 {
	if hoe.prod.End == nil {
		limit := make([]int, len(hoe.X0))
		for i := range limit {
			limit[i] = 2
		}
		hoe.prod = pfutil.NewProduct(limit)
	}

	comb := hoe.prod.Next()
	if comb == nil {
		return nil
	}

	nextPoint := make([]float64, len(hoe.X0))
	for i := range nextPoint {
		nextPoint[i] = (2.0*float64(comb[i]) - 1.0) * hoe.X0[i]
	}
	return nextPoint
}

// OnMinimizationFinished does nothing (included to satisfy the InitialGuessExplorer interface)
func (hoe *HyperOctantExplorer) OnMinimizationFinished(res *optimize.Result) {}

// SinglePointExplorer explores single initial guess
type SinglePointExplorer struct {
	X0         []float64
	nextCalled bool
}

// Next returns X0 the first time it is called. Subsequent calls will
// return nil
func (spe *SinglePointExplorer) Next() []float64 {
	if !spe.nextCalled {
		spe.nextCalled = true
		return spe.X0
	}
	return nil
}

// OnMinimizationFinished does nothing (included to satisfy the InitialGuessExplorer interface)
func (spe *SinglePointExplorer) OnMinimizationFinished(res *optimize.Result) {}
