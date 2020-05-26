package pfutil

// nodeIdx2 returns the node index for 2D grid
func nodeIdx2(domainSize []int, idx []int) int {
	return idx[0]*domainSize[1] + idx[1]
}

// nodeIdx3 returns the node index for 3D grid
func nodeIdx3(domainSize []int, idx []int) int {
	return idx[2]*domainSize[0]*domainSize[1] + idx[0]*domainSize[1] + idx[1]
}

// NodeIdx returns the index of the node corresponding to a given typle of index
func NodeIdx(domainSize []int, idx []int) int {
	if len(domainSize) == 2 && len(idx) == 2 {
		return nodeIdx2(domainSize, idx)
	} else if len(domainSize) == 3 && len(idx) == 3 {
		return nodeIdx3(domainSize, idx)
	}
	panic("util: Domain size and idx has to be of length 2 or 3")
}

func pos3(domainSize []int, nodeNum int) []int {
	col := nodeNum % domainSize[1]
	row := (nodeNum / domainSize[1]) % domainSize[0]
	depth := nodeNum / (domainSize[0] * domainSize[1])
	return []int{row, col, depth}
}

func pos2(domainSize []int, nodeNum int) []int {
	col := nodeNum % domainSize[1]
	row := nodeNum / domainSize[1]
	return []int{row, col}
}

// Pos converts the node number to position
func Pos(domainSize []int, nodeNum int) []int {
	if len(domainSize) == 2 {
		return pos2(domainSize, nodeNum)
	} else if len(domainSize) == 3 {
		return pos3(domainSize, nodeNum)
	}
	panic("util: Domain size has to be either 2 or 3")
}
