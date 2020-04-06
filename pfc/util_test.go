package pfc

import (
	"testing"
)

func TestProduct(t *testing.T) {
	for i, test := range []struct {
		End    []int
		Expect [][]int
	}{
		{
			End: []int{2, 2, 2},
			Expect: [][]int{
				[]int{0, 0, 0},
				[]int{0, 0, 1},
				[]int{0, 1, 0},
				[]int{0, 1, 1},
				[]int{1, 0, 0},
				[]int{1, 0, 1},
				[]int{1, 1, 0},
				[]int{1, 1, 1},
			},
		},
		{
			End: []int{2, 3},
			Expect: [][]int{
				[]int{0, 0},
				[]int{0, 1},
				[]int{0, 2},
				[]int{1, 0},
				[]int{1, 1},
				[]int{1, 2},
			},
		},
	} {
		prod := NewProduct(test.End)
		counter := 0
		for current := prod.Next(); current != nil; current = prod.Next() {
			for j := range current {
				if current[j] != test.Expect[counter][j] {
					t.Errorf("Test #%d: Expected %v got %v\n", i, test.Expect[counter], current)
					return
				}
			}
			counter++
		}
		if counter != len(test.Expect) {
			t.Errorf("Loop did not take place")
		}
	}
}
