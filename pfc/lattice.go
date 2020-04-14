package pfc

import (
	"math"
	"sort"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/spatial/r3"
)

// Cell is a type used to represent a unit cell. In the underlying CellVec
// matrix, each column represents a cell vector
type Cell struct {
	CellVec *mat.Dense
}

// Volume returns the volume of the cell. For 2D lattices this would
// be the surface area
func (c *Cell) Volume() float64 {
	return mat.Det(c.CellVec)
}

// ReciprocalCell is a type used to represent the reciprocal lattice. In underlying CellVec
// matrix, each column represents a cell vector. The easiest way to initialize this type is
// via the Reciprocal method of the Cell type.
type ReciprocalCell struct {
	CellVec *mat.Dense
}

// HKLVector returns a vector correslonding to the passed miller indices
func (rc *ReciprocalCell) HKLVector(miller Miller) []float64 {
	_, col := rc.CellVec.Dims()
	res := make([]float64, col)
	for i := 0; i < col; i++ {
		res[i] = float64(miller.H)*rc.CellVec.At(i, 0) + float64(miller.K)*rc.CellVec.At(i, 1)
		if col == 3 {
			res[i] += float64(miller.L) * rc.CellVec.At(i, 2)
		}
	}
	return res
}

// CellChange returns the change in the reciprocal lattice resulting from a
// given change in the real space lattice. It is assumed that the change is small
// and the function relies on a first order series expansion of the inverse real
// space cell
func (rc *ReciprocalCell) CellChange(realSpaceChange *mat.Dense) *mat.Dense {
	r, c := realSpaceChange.Dims()
	change := mat.NewDense(r, c, nil)
	change.Product(rc.CellVec, realSpaceChange, rc.CellVec)
	change.Scale(-1.0/(2.0*math.Pi), change)
	return change
}

// ChangeHKLVector returns the change in the HKL vector originating from a
// change in the real space lattice
func (rc *ReciprocalCell) ChangeHKLVector(miller Miller, realSpaceChange *mat.Dense) []float64 {
	change := rc.CellChange(realSpaceChange)
	_, col := change.Dims()

	res := make([]float64, col)
	for i := 0; i < col; i++ {
		res[i] = float64(miller.H)*change.At(i, 0) + float64(miller.K)*change.At(i, 1)
		if col == 3 {
			res[i] += float64(miller.L) * change.At(i, 2)
		}
	}
	return res
}

// Reciprocal returns a cell object representing the reciprocal cell
// The underlying matrix of the reciprocal lattice is given by
//
// M^T = 2*pi*C^{-1}
//
// where each column of M is a reciprocal lattice vector and each column
// if C is a real lattice vector
func (c *Cell) Reciprocal() ReciprocalCell {
	row, col := c.CellVec.Dims()
	rCell := ReciprocalCell{
		CellVec: mat.NewDense(row, col, nil),
	}
	tmpCell := mat.NewDense(row, col, nil)
	tmpCell.Inverse(c.CellVec)
	for i := 0; i < row; i++ {
		for j := 0; j < col; j++ {
			rCell.CellVec.Set(j, i, 2.0*math.Pi*tmpCell.At(i, j))
		}
	}
	return rCell
}

// FCC construct a FCC cell
func FCC(a float64) Cell {
	cell := Cell{
		CellVec: mat.NewDense(3, 3, nil),
	}
	cell.CellVec.Set(0, 0, a/2.0)
	cell.CellVec.Set(1, 0, a/2.0)

	cell.CellVec.Set(1, 1, a/2.0)
	cell.CellVec.Set(2, 1, a/2.0)

	cell.CellVec.Set(0, 2, a/2.0)
	cell.CellVec.Set(2, 2, a/2.0)
	return cell
}

// SC3D returns the primitive unit cell for a simple cubic structure
// with lattice parameter a in 3D
func SC3D(a float64) Cell {
	cell := Cell{
		CellVec: mat.NewDense(3, 3, nil),
	}
	cell.CellVec.Set(0, 0, a)
	cell.CellVec.Set(1, 1, a)
	cell.CellVec.Set(2, 2, a)
	return cell
}

// BCC returns a primitive cell for the BCC structure
func BCC(a float64) Cell {
	cell := Cell{
		CellVec: mat.NewDense(3, 3, nil),
	}
	cell.CellVec.Set(0, 0, a/2.0)
	cell.CellVec.Set(1, 0, a/2.0)
	cell.CellVec.Set(2, 0, a/2.0)

	cell.CellVec.Set(0, 1, -a/2.0)
	cell.CellVec.Set(1, 1, a/2.0)
	cell.CellVec.Set(2, 1, a/2.0)

	cell.CellVec.Set(0, 2, a/2.0)
	cell.CellVec.Set(1, 2, -a/2.0)
	cell.CellVec.Set(2, 2, a/2.0)
	return cell
}

// SC2D returns the primitive cell of a 2D cubical system
func SC2D(a float64) Cell {
	cell := Cell{
		CellVec: mat.NewDense(2, 2, nil),
	}
	cell.CellVec.Set(0, 0, a)
	cell.CellVec.Set(1, 1, a)
	return cell
}

// Triangular2D returns the primitive cell of a triangular
// lattice in 2D
func Triangular2D(a float64) Cell {
	cell := Cell{
		CellVec: mat.NewDense(2, 2, nil),
	}

	cell.CellVec.Set(0, 0, a)
	cell.CellVec.Set(0, 1, a/2.0)
	cell.CellVec.Set(1, 1, a/2.0)
	return cell
}

// CubicUnitCellDensity returns the density of a plane when the underlying
// Bravais lattice is cubic.
func CubicUnitCellDensity(miller Miller) float64 {
	millerArray := make([]int, 3)
	millerArray[0] = miller.At(0)
	millerArray[1] = miller.At(1)
	millerArray[2] = miller.At(2)
	sort.Ints(millerArray)

	if millerArray[0] == 0 && millerArray[1] == 0 {
		return 1.0
	}
	v1 := r3.Vec{}
	v2 := r3.Vec{}
	v3 := r3.Vec{}

	if millerArray[0] == 1 {
		v1.X = 1.0
		v2.Y = 1.0 / float64(millerArray[1])
		v3.Z = 1.0 / float64(millerArray[2])
		v3 = v3.Sub(v1) // v3 --> v3 - v1
		v2 = v2.Sub(v1) // v2 --> v2 - v1
	} else if millerArray[0] == 0 {
		v2.Y = 1.0 / float64(millerArray[1])
		v3.Z = 1.0 / float64(millerArray[2])
		v3 = v3.Sub(v2)
		v2.Y = 0.0
		v2.X = 1.0
	} else {
		panic("The smallest Miller index has to be 0 or 1")
	}

	// TODO: v0.7 of gonum does not include Cross, but it is included in development version
	// Use the line below after next release of gonum
	// cross := v1.Cross(v2)
	cross := r3.Vec{
		X: v3.Y*v2.Z - v3.Z*v2.Y,
		Y: v3.Z*v2.X - v3.X*v2.Z,
		Z: v3.X*v2.Y - v3.Y*v2.X,
	}
	area := math.Sqrt(cross.X*cross.X + cross.Y*cross.Y + cross.Z*cross.Z)
	return 1.0 / area
}
