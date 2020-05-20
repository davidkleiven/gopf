package pfutil

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
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
		Data: make([]float64, ProdInt(dims)),
	}
}

// Copy returns a copy of the grid
func (g Grid) Copy() Grid {
	newGrid := Grid{
		Dims: make([]int, len(g.Dims)),
		Data: make([]float64, len(g.Data)),
	}
	copy(newGrid.Dims, g.Dims)
	copy(newGrid.Data, g.Data)
	return newGrid
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

// FromComplex extracts the real part of a complex array
func (g *Grid) FromComplex(carray []complex128) {
	for i := range carray {
		g.Data[i] = real(carray[i])
	}
}

// SaveCsv stores the grid in a text file format. The format of the
// produced file is
// 1. For 2D grids
// x, y, value
// 2. For 3D grids
// x, y, z, value
func (g Grid) SaveCsv(fname string) {
	f, err := os.Create(fname)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	defer f.Close()

	header := "x,y,value\n"
	dim := len(g.Dims)
	if dim == 3 {
		header = "x,y,z,value\n"
	}
	f.WriteString(header)
	pos := make([]int, 3)
	for i := range g.Data {
		g.Pos(i, pos)
		posString := []string{}
		for j := 0; j < dim; j++ {
			posString = append(posString, strconv.Itoa(pos[j]))
		}
		text := strings.Join(posString, ",")
		text += fmt.Sprintf(",%f\n", g.Data[i])
		f.WriteString(text)
	}
}

// Rotate2D rotates 2D grids about the center. Angle is the rotation angle in radians.
func (g *Grid) Rotate2D(angle float64) {
	if len(g.Dims) != 2 {
		panic("Grid: Rotate2D: Grid must be 2D")
	}
	srgImg := make([]float64, len(g.Data))
	copy(srgImg, g.Data)

	// Negative angles because we loop over the destination image
	// and fill it with a pixel in the source image
	cosa, sina := math.Cos(-angle), math.Sin(-angle)
	pos := make([]int, 2)
	rotPos := make([]int, 2)
	for i := range g.Data {
		g.pos2d(i, pos)
		pos[0] -= g.Dims[0] / 2
		pos[1] -= g.Dims[1] / 2

		// Find rotated pixel in the the source image
		rotPos[0] = int(float64(pos[0])*cosa + float64(pos[1])*sina)
		rotPos[1] = int(float64(pos[1])*cosa - float64(pos[0])*sina)
		rotPos[0] += g.Dims[0] / 2
		rotPos[1] += g.Dims[1] / 2
		Wrap(rotPos, g.Dims)
		g.Data[i] = srgImg[g.index2d(rotPos)]
	}
}
