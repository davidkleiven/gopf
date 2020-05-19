package pf

import (
	"math"
	"math/cmplx"

	"github.com/davidkleiven/gopf/pfutil"
)

// TimeDepSource is a generic interface for functions that can be used as sources
type TimeDepSource func(t float64) float64

// Source is structure used to add source terms to an equation
type Source struct {
	Pos []float64
	f   TimeDepSource
}

// NewSource returns a new source instance
func NewSource(pos []float64, f TimeDepSource) Source {
	return Source{Pos: pos, f: f}
}

// Eval evaluates the fourier transformed source term and places the result in data
func (s *Source) Eval(freq Frequency, t float64, data []complex128) {
	for i := range data {
		k := freq(i)
		data[i] = complex(s.f(t), 0.0) * cmplx.Exp(-complex(0.0, 2.0*math.Pi*pfutil.Dot(k, s.Pos)))
	}
}

// Sources is a type used to represent many sources
type Sources []Source
