package pfutil

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// Affine is a type used to represent affine transformations. The transformation
// itself is stored as a 4x4 matrix
type Affine struct {
	Mat *mat.Dense
}

// Apply applies the transformation to the passed position vector.
func (a *Affine) Apply(pos []float64) {
	vec := mat.NewVecDense(4, nil)
	for i := range pos {
		vec.SetVec(i, pos[i])
	}
	vec.SetVec(3, 1.0)
	vec.MulVec(a.Mat, vec)
	for i := range pos {
		pos[i] = vec.AtVec(i)
	}
}

// Append applies the passed transformation after the current. If the receiver's
// transformation is A and other's transformation matrix is B, the updated transformation
// matrix is given by A <- B*A
func (a *Affine) Append(other Affine) {
	a.Mat.Mul(other.Mat, a.Mat)
}

// NewAffine initializes a new empty transformation
func NewAffine() Affine {
	return Affine{
		Mat: mat.NewDense(4, 4, nil),
	}
}

// Identity initializes a new identity transformation
func Identity() Affine {
	affine := NewAffine()
	for i := 0; i < 4; i++ {
		affine.Mat.Set(i, i, 1.0)
	}
	return affine
}

// Translation initializes a new translation transformation
func Translation(vec []float64) Affine {
	affine := Identity()
	for i := range vec {
		affine.Mat.Set(i, 3, vec[i])
	}
	affine.Mat.Set(3, 3, 1.0)
	return affine
}

// Scaling returns a new scaling transformation
func Scaling(vec []float64) Affine {
	affine := NewAffine()
	for i := range vec {
		affine.Mat.Set(i, i, vec[i])
	}
	affine.Mat.Set(3, 3, 1.0)
	return affine
}

// RotZ returns the rotation transformation corresonding to rotation about the z-axis by
// an angle alpha. c = cos alpha, s = sin alpha, the matrix is
// **            **
// * c    -s    0 *
// * s     c    0 *
// * 0     0    1 *
// **            **
func RotZ(angle float64) Affine {
	affine := NewAffine()
	c := math.Cos(angle)
	s := math.Sin(angle)

	affine.Mat.Set(0, 0, c)
	affine.Mat.Set(0, 1, -s)
	affine.Mat.Set(1, 0, s)
	affine.Mat.Set(1, 1, c)
	affine.Mat.Set(2, 2, 1.0)
	affine.Mat.Set(3, 3, 1.0)
	return affine
}

// RotY the rotation transformation an angle about the y-axis
// c = cos angle, s = sin angle
// **           **
// * c    0    s *
// * 0    1    0 *
// * -s   0    c *
// **           **
func RotY(angle float64) Affine {
	affine := NewAffine()
	c := math.Cos(angle)
	s := math.Sin(angle)

	affine.Mat.Set(0, 0, c)
	affine.Mat.Set(0, 2, s)
	affine.Mat.Set(1, 1, 1.0)
	affine.Mat.Set(2, 0, -s)
	affine.Mat.Set(2, 2, c)
	affine.Mat.Set(3, 3, 1.0)
	return affine
}

// RotX returns the rotation transformation corresponding to a rotation about th x-axis
// c = cos angle, s = sin angle
// **           **
// * 1    0    0 *
// * 0    c   -s *
// * 0    s    c *
// **           **
func RotX(angle float64) Affine {
	affine := NewAffine()
	c := math.Cos(angle)
	s := math.Sin(angle)

	affine.Mat.Set(0, 0, 1.0)
	affine.Mat.Set(1, 1, c)
	affine.Mat.Set(1, 2, -s)
	affine.Mat.Set(2, 1, s)
	affine.Mat.Set(2, 2, c)
	affine.Mat.Set(3, 3, 1.0)
	return affine
}
