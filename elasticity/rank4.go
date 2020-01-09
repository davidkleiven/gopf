package elasticity

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

// Rank4 is a type used to represent a 3x3x3x3 tensor
type Rank4 struct {
	Data []float64
}

// NewRank4 returns a new rank4 tensor
func NewRank4() Rank4 {
	return Rank4{
		Data: make([]float64, 81),
	}
}

// index returns the index corresponding to (i, j, k l)
func (r *Rank4) index(i, j, k, l int) int {
	return i*27 + j*9 + k*3 + l
}

// At returns the element at position (i, j, k, l)
func (r *Rank4) At(i, j, k, l int) float64 {
	return r.Data[r.index(i, j, k, l)]
}

// Set sets a new value at position (i, j, k, l)
func (r *Rank4) Set(i, j, k, l int, v float64) {
	r.Data[r.index(i, j, k, l)] = v
}

// Rotate rotates the tensor
func (r *Rank4) Rotate(rot mat.Matrix) {
	rotated := make([]float64, 81)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					// Inner indices
					for m := 0; m < 3; m++ {
						for n := 0; n < 3; n++ {
							for p := 0; p < 3; p++ {
								for q := 0; q < 3; q++ {
									idx := r.index(i, j, k, l)
									rotated[idx] += rot.At(i, m) * rot.At(j, n) * rot.At(k, p) * rot.At(l, q) * r.At(m, n, p, q)
								}
							}
						}
					}
				}
			}
		}
	}
	copy(r.Data, rotated)
}

// ContractLast contracts the two last indices with the passed matrix
func (r *Rank4) ContractLast(tensor *mat.Dense) *mat.Dense {
	out := mat.NewDense(3, 3, nil)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					cur := out.At(i, j)
					out.Set(i, j, cur+r.At(i, j, k, l)*tensor.At(k, l))
				}
			}
		}
	}
	return out
}

// Shear modulus returns the shear modulus when the bulk modulus
// and poisson ratio is known
func Shear(bulkMod float64, poisson float64) float64 {
	return 3.0 * bulkMod * (1.0 - 2.0*poisson) / (2.0 * (1.0 + poisson))
}

// Isotropic returns an elastic tensor
func Isotropic(bulkMod float64, poisson float64) Rank4 {
	shear := Shear(bulkMod, poisson)
	tensor := NewRank4()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				for l := 0; l < 3; l++ {
					value := 0.0
					if i == j && k == l {
						value += bulkMod - 2.0*shear/3.0
					}

					if i == k && j == l {
						value += shear
					}

					if i == l && j == k {
						value += shear
					}
					tensor.Set(i, j, k, l, value)
				}
			}
		}
	}
	return tensor
}

// CubicMaterial returns the elastic tensor for a cubic material, where
// c11, c12 and c44 are constants in the Voigt representation
func CubicMaterial(c11 float64, c12 float64, c44 float64) Rank4 {
	tensor := NewRank4()
	for i := 0; i < 3; i++ {
		tensor.Set(i, i, i, i, c11)
	}
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 3; j++ {
			tensor.Set(i, i, j, j, c12)
			tensor.Set(j, j, i, i, c12)

			tensor.Set(i, j, i, j, c44)
			tensor.Set(j, i, j, i, c44)
		}
	}
	return tensor
}

// RotationMatrix creates the rotation matrix corresponding to a
// rotation around the specified axis
func RotationMatrix(angle float64, axis int) *mat.Dense {
	c := math.Cos(angle)
	s := math.Sin(angle)
	rot := mat.NewDense(3, 3, nil)
	if axis == 0 {
		rot.Set(0, 0, 1.0)
		rot.Set(1, 1, c)
		rot.Set(2, 2, c)
		rot.Set(1, 2, s)
		rot.Set(2, 1, -s)
	} else if axis == 1 {
		rot.Set(1, 1, 1.0)
		rot.Set(0, 0, c)
		rot.Set(2, 2, c)
		rot.Set(0, 2, s)
		rot.Set(2, 0, -s)
	} else if axis == 2 {
		rot.Set(2, 2, 1.0)
		rot.Set(0, 0, c)
		rot.Set(1, 1, c)
		rot.Set(0, 1, s)
		rot.Set(1, 0, -s)
	}
	return rot
}
