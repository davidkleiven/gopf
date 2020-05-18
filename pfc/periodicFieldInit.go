package pfc

import (
	"math"

	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/gonum/mat"
)

// UnitCell is a type that represents a unit cell of a crystal structure
type UnitCell struct {
	// Cell represents where each column is a cell vector
	Cell Cell

	// Basis represents the positions inside the unit cell.
	// It is assumed that the basis is given in scaled coordinates.
	// Cartesian coordinates can thus be obtained by C = Cell.dot(Basis)
	Basis *mat.Dense
}

// AtomicKernel is a function type that returns a density
type AtomicKernel interface {
	// Eval evaluates the kernel at the given coordinate
	Eval(x float64) float64

	// Cutoff specifies a maximum value beyond which Eval returns
	// essentially zero
	Cutoff() float64
}

// CornersScaledCrd returns the positions of the corners in scaled coordinates
// Cell represents the unit cell and domainSize gives the size of the domain
// which is assumed to be the box [0, domainSize[0]] x [0, domainSize[1]]
// (x [0, domainSize[2]] if in 3D)
func CornersScaledCrd(cell Cell, domainSize []int) *mat.Dense {
	dim, _ := cell.CellVec.Dims()
	numCorners := 4
	end := make([]int, 3)
	end[0] = 2
	end[1] = 2
	end[2] = 2
	if dim == 3 {
		numCorners = 8
	}

	cornerLoc := mat.NewDense(dim, numCorners, nil)
	prod := pfutil.NewProduct(end[:dim])
	counter := 0
	vec := make([]float64, 3)
	for idx := prod.Next(); idx != nil; idx = prod.Next() {
		for i := 0; i < dim; i++ {
			vec[i] = float64(domainSize[i] * idx[i])
		}

		// Transfer to the corner matrix
		for i := 0; i < dim; i++ {
			cornerLoc.Set(i, counter, vec[i])
		}
		counter++
	}

	// Find the location of the corners in scaled coordinates
	cornerLoc.Solve(cell.CellVec, cornerLoc)
	return cornerLoc
}

// isInsideBox returns true if a point is inside the box specified by domainSize
// (e.g. 0 <= point.AtVec(i) && point.AtVec(i) + threshold < domainSize[i]) where
// i is 0, 1 (and 2 if 3D)
func isInsideBox(point *mat.VecDense, domainSize []int, threshold float64) bool {
	for i := range domainSize {
		if (point.AtVec(i) < 0.0) || (point.AtVec(i)+threshold > float64(domainSize[i])) {
			return false
		}
	}
	return true
}

// GaussianKernel is a type that represents the field from one atoms as a Gaussian
// centered at position of the atom
type GaussianKernel struct {
	// Width is the "standard deviation" of the gaussian
	Width float64
}

// Cutoff returns a cutoff where the kernel is essentiall zero
func (gk *GaussianKernel) Cutoff() float64 {
	return 3.0 * gk.Width
}

// Eval evaluates the gaussian kernel
func (gk *GaussianKernel) Eval(x float64) float64 {
	exponent := math.Pow(x/gk.Width, 2)
	return math.Exp(-exponent / 2.0)
}

// CircleKernel represents the field from one atom as a square
type CircleKernel struct {
	Radius float64
}

// Cutoff returns the width of the square
func (ck *CircleKernel) Cutoff() float64 {
	return ck.Radius
}

// Eval return 1 if -width/2 <= x < width/2, and zero otherwise
func (ck *CircleKernel) Eval(x float64) float64 {
	if x >= 0.0 && x < ck.Radius+1.0 {
		return 1.0
	}
	return 0.0
}

// Area returns the of the circle
func (ck *CircleKernel) Area() float64 {
	return math.Pi * ck.Radius * ck.Radius
}

// BuildCrystal constructs a field that is organized as a crystal
// ucell is the unit cell, kernel is the functional form used to represent
// the field form one atom and grid is a computational grid that will be
// assigned the field corresponding to the crystal structure
func BuildCrystal(ucell UnitCell, kernel AtomicKernel, grid *pfutil.Grid) {
	corners := CornersScaledCrd(ucell.Cell, grid.Dims)
	minval := corners.At(0, 0)
	maxval := corners.At(0, 0)
	data := corners.RawMatrix().Data
	for i := range data {
		if data[i] < minval {
			minval = data[i]
		}
		if data[i] > maxval {
			maxval = data[i]
		}
	}

	dims, _ := ucell.Cell.CellVec.Dims()
	end := make([]int, dims)
	rng := maxval - minval
	for i := range end {
		end[i] = int(rng) + 1
	}
	prod := pfutil.NewProduct(end)
	origin := mat.NewVecDense(dims, nil)
	_, numBasis := ucell.Basis.Dims()
	eucledianPos := mat.NewDense(dims, numBasis, nil)

	cutoff := kernel.Cutoff()
	cutoffEnd := make([]int, dims)
	for j := range cutoffEnd {
		cutoffEnd[j] = int(2.0 * cutoff)
	}

	shiftedPos := make([]int, dims)
	vol := ucell.Cell.Volume()
	L := math.Pow(vol, 1.0/3.0) / 2.0

	for idx := prod.Next(); idx != nil; idx = prod.Next() {
		for i := range idx {
			origin.SetVec(i, float64(idx[i])+minval)
		}
		origin.MulVec(ucell.Cell.CellVec, origin)
		if !isInsideBox(origin, grid.Dims, L) {
			continue
		}
		eucledianPos.Mul(ucell.Cell.CellVec, ucell.Basis)

		for j := 0; j < numBasis; j++ {
			nodeIter := pfutil.NewProduct(cutoffEnd)
			// Iterator over all nodes within the cutoff
			for pos := nodeIter.Next(); pos != nil; pos = nodeIter.Next() {
				distSq := 0.0
				for k := 0; k < dims; k++ {
					delta := float64(pos[k]) - cutoff + 0.5
					shiftedPos[k] = int(delta + origin.AtVec(k) + eucledianPos.At(k, j))
					distSq += math.Pow(delta, 2)
				}
				pfutil.Wrap(shiftedPos, grid.Dims)

				// Update the field
				grid.Set(shiftedPos, kernel.Eval(math.Sqrt(distSq)))
			}
		}
	}
}
