package pfc

import "gonum.org/v1/gonum/integrate/quad"

// QuadSquare calculates the 2D integral over a unit square [0, 1] x [0, 1]
// using tensorial Legendre quadrature. The tensorial Legendre quadrature,
// is not optimal with respect to function evaluations, but does the job.
func QuadSquare(f func(float64, float64) float64, n int) float64 {
	I := 0.0
	rule := quad.Legendre{}
	for i := 0; i < n; i++ {
		x, wx := rule.FixedLocationSingle(n, i, 0.0, 1.0)
		for j := 0; j < n; j++ {
			y, wy := rule.FixedLocationSingle(n, j, 0.0, 1.0)
			I += wx * wy * f(x, y)
		}
	}
	return I
}

// QuadCube calculates the 3D integral over a unit cube [0, 1] x [0, 1] x [0, 1]
// using tensorial Legendre quadrature. The tensorial Legendre quadrature,
// is not optimal with respect to function evaluations, but does the job.
func QuadCube(f func(float64, float64, float64) float64, n int) float64 {
	I := 0.0
	rule := quad.Legendre{}
	for i := 0; i < n; i++ {
		x, wx := rule.FixedLocationSingle(n, i, 0.0, 1.0)
		for j := 0; j < n; j++ {
			y, wy := rule.FixedLocationSingle(n, j, 0.0, 1.0)
			for k := 0; k < n; k++ {
				z, wz := rule.FixedLocationSingle(n, k, 0.0, 1.0)
				I += wx * wy * wz * f(x, y, z)
			}
		}
	}
	return I
}
