package pfutil

import (
	"testing"
)

func TestVoronoi(t *testing.T) {
	domainSize := []int{8, 8}
	data := make([]int, 64)
	locations := [][]int{
		{4, 4},
		{7, 7},
		{1, 1},
	}
	pts := make([]int, len(locations))
	for i := range pts {
		pts[i] = NodeIdx(domainSize, locations[i])
	}
	Voronoi(pts, data, domainSize)
	expect := [][]int{
		{1, 2, 2, 2, 1, 1, 1, 1},
		{2, 2, 2, 2, 0, 1, 1, 1},
		{2, 2, 2, 0, 0, 0, 0, 2},
		{2, 2, 0, 0, 0, 0, 0, 2},
		{1, 0, 0, 0, 0, 0, 0, 0},
		{1, 1, 0, 0, 0, 0, 0, 1},
		{1, 1, 0, 0, 0, 0, 1, 1},
		{1, 1, 2, 2, 0, 1, 1, 1},
	}
	for i := range data {
		pos := Pos(domainSize, i)
		e := expect[pos[0]][pos[1]]
		if e != data[i] {
			t.Errorf("Expected %d got %d\n", e, data[i])
		}
	}
}
