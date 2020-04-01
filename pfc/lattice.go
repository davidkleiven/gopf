package pfc

import (
	"math"

	"gonum.org/v1/gonum/mat"
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
