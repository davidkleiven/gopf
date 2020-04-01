package pfc

// Miller is a struct used to represent the Miller indices of a crystal plane
type Miller struct {
	H, K, L int
}

// Equal checks if two miller indices are equal
func (m *Miller) Equal(m2 Miller) bool {
	return m.H == m2.H && m.K == m2.K && m.L == m2.L
}

// Return the Miller index as if it was an array. H is the zeroth element
// K is the first element and L is the last element
func (m *Miller) At(i int) int {
	if i == 0 {
		return m.H
	} else if i == 1 {
		return m.K
	}
	return m.L
}

// factorial is a helper that calculates the factorial of n
func factorial(n int) int {
	f := 1
	for i := 1; i <= n; i++ {
		f *= i
	}
	return f
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// NumEquivalent3D returns the number of equivalent planes
func NumEquivalent3D(miller Miller) int {
	numEq := 1

	// Take into account the different combinations of +- the index
	if miller.H != 0 {
		numEq *= 2
	}

	if miller.K != 0 {
		numEq *= 2
	}

	if miller.L != 0 {
		numEq *= 2
	}

	// Take into account all different permutations
	numEqual := 1
	if abs(miller.K) == abs(miller.H) {
		numEqual++
	}
	if abs(miller.L) == abs(miller.K) {
		numEqual++
	}

	numPerm := factorial(3) / factorial(numEqual)
	return numEq * numPerm
}

// NumEquivalent2D returns the number of equivalent miller indices in 2D
func NumEquivalent2D(miller Miller) int {
	num := 0
	for _, equiv := range EquivalentMiller(miller) {
		if equiv.L == 0 {
			num++
		}
	}
	return num
}

// NumEquivalent returns the number of equivalent indices
func NumEquivalent(miller Miller, dim int) int {
	switch dim {
	case 2:
		return NumEquivalent2D(miller)
	case 3:
		return NumEquivalent3D(miller)
	default:
		panic("Dimension has to be either 2 or 3")
	}
}

// EquivalentMiller returns a array with all miller indices that
// are equivalent to the one passed (including the one passed)
func EquivalentMiller(miller Miller) []Miller {
	res := make([]Miller, 8*6)
	permutations := [][]int{
		[]int{0, 1, 2},
		[]int{0, 2, 1},
		[]int{2, 0, 1},
		[]int{2, 1, 0},
		[]int{1, 2, 0},
		[]int{1, 0, 2},
	}
	millerArray := []int{miller.H, miller.K, miller.L}
	current := 0
	for _, p := range permutations {
		m := Miller{
			H: millerArray[p[0]],
			K: millerArray[p[1]],
			L: millerArray[p[2]],
		}
		newEquiv := EquivalentMillerNoPermutations(m)
		copy(res[current:], newEquiv)
		current += len(newEquiv)
	}
	res = res[:current]
	// Remove all equal miller indices
	end := len(res)
	i := 0
	for i < end {
		j := i + 1
		for j < end {
			if res[i].Equal(res[j]) {
				res[j], res[end-1] = res[end-1], res[j]
				end--
			} else {
				j++
			}
		}
		i++
	}
	return res[:end]
}

// EquivalentMillerNoPermutations returns all equivalent miller indices
// without taking permutations into account
func EquivalentMillerNoPermutations(miller Miller) []Miller {
	res := make([]Miller, 8)

	counter := 0
	maxH := 1
	if miller.H == 0 {
		maxH = 0
	}

	maxK := 1
	if miller.K == 0 {
		maxK = 0
	}

	maxL := 1
	if miller.L == 0 {
		maxL = 0
	}

	for sgnH := -1; sgnH <= maxH; sgnH += 2 {
		for sgnK := -1; sgnK <= maxK; sgnK += 2 {
			for sgnL := -1; sgnL <= maxL; sgnL += 2 {
				res[counter] = Miller{
					H: sgnH * miller.H,
					K: sgnK * miller.K,
					L: sgnL * miller.L,
				}
				counter++
			}
		}
	}
	return res[:counter]
}
