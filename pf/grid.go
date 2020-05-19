package pf

import (
	"fmt"

	"github.com/davidkleiven/gopf/pfutil"
)

// Grid is a type that represents a computatoinal grid
type Grid struct {
	Dims []int
	Data []float64
}

// NewGrid initializes a new grid type
func NewGrid(dims []int) Grid {
	return Grid{
		Dims: dims,
		Data: make([]float64, pfutil.ProdInt(dims)),
	}
}

func (g Grid) index2d(pos []int) int {
	return pos[0]*g.Dims[1] + pos[1]
}

func (g Grid) index3d(pos []int) int {
	return pos[2]*g.Dims[0]*g.Dims[1] + pos[0]*g.Dims[1] + pos[1]
}

// index returnds the underlying index corresponding to the point
func (g Grid) index(pos []int) int {
	if len(pos) == 2 {
		return g.index2d(pos)
	}
	return g.index3d(pos)
}

// pos2d returns position in 2D
func (g Grid) pos2d(i int, pos []int) {
	pos[0] = i / g.Dims[1]
	pos[1] = i % g.Dims[1]
}

// pos3D returns the position i 3D
func (g Grid) pos3d(i int, pos []int) {
	fmt.Printf("%v\n", g.Dims)
	pos[0] = (i / g.Dims[1]) % g.Dims[0]
	pos[1] = i % g.Dims[1]
	pos[2] = i / (g.Dims[0] * g.Dims[1])
}

// Pos returns the positions that corresponds to index i
func (g Grid) Pos(i int, pos []int) {
	if len(g.Dims) == 2 {
		g.pos2d(i, pos)
	}
	g.pos3d(i, pos)
}

// Set sets a value
func (g *Grid) Set(pos []int, value float64) {
	g.Data[g.index(pos)] = value
}

// Get gets a value at the given position
func (g Grid) Get(pos []int) float64 {
	return g.Data[g.index(pos)]
}

// ToComplex converts the underlying data array to a complex array.
// The data is inserted as the real part and the imaginary part is
// zero.
func (g Grid) ToComplex() []complex128 {
	carray := make([]complex128, len(g.Data))
	for i := range g.Data {
		carray[i] = complex(g.Data[i], 0.0)
	}
	return carray
}
