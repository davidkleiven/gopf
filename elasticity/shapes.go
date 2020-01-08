package elasticity

import (
	"math"

	"github.com/davidkleiven/gosfft/sfft"
)

// Ellipsoid returns a voxel representation of an ellipsoid with half-axes given by
// a, b and c. The overall domai size will be N x N x N
func Ellipsoid(N int, a float64, b float64, c float64) *sfft.CMat3 {
	voxels := sfft.NewCMat3(N, N, N, nil)
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			for k := 0; k < N; k++ {
				x := float64(i - N/2)
				y := float64(j - N/2)
				z := float64(k - N/2)
				value := math.Pow(x/a, 2) + math.Pow(y/b, 2) + math.Pow(z/c, 2)
				if value <= 1.0 {
					voxels.Set(i, j, k, complex(1.0, 0.0))
				}
			}
		}
	}
	return voxels
}
