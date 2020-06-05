package pfutil

import "math"

// Voronoi assigns each pixel/voxel of the data array to the closest
// points amongst the generating points in pts. Domain size is an array
// that specifies the sizes in X, Y (and Z iif 3D) direction
func Voronoi(pts []int, data []int, domainSize []int) {
	for i := range data {
		data[i] = closestPoint(pts, i, domainSize)
	}
}

// pointDistSq returns the squared distance between two points'
// minimum image convention is used
func pointDistSq(i int, j int, domainSize []int) int {
	ipos := Pos(domainSize, i)
	jpos := Pos(domainSize, j)

	dist := 0
	for k := range ipos {
		smallestDiff := ipos[k] - jpos[k]
		for _, shift := range []int{-1, 1} {
			diff := ipos[k] - jpos[k] + shift*domainSize[k]
			if absInt(diff) < absInt(smallestDiff) {
				smallestDiff = diff
			}
		}
		dist += smallestDiff * smallestDiff
	}
	return dist
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func closestPoint(pts []int, point int, domainSize []int) int {
	mindistSq := math.MaxInt32
	bestPoint := 0
	for i := range pts {
		distSq := pointDistSq(pts[i], point, domainSize)
		if distSq < mindistSq {
			mindistSq = distSq
			bestPoint = i
		}
	}
	return bestPoint
}
