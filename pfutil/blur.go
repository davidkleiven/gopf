package pfutil

// BlurKernel is an interfae for kernel functions used for blurring
type BlurKernel interface {
	// Value returns the value of the kernel
	Value(x []int) float64

	// Cutoff returns a value beyond which the kernel is zero
	// (e.g. the kernel is assumed to be zero outisde the box
	// -Cutoff() < x < Cutoff() )
	Cutoff() int
}

// BoxKernel returns a blurring kernel that is 1 inside a box and zero outside
type BoxKernel struct {
	Width int
}

// Cutoff the half-width of the box
func (bk *BoxKernel) Cutoff() int {
	return bk.Width/2 + 1
}

// Value returns the box kernel
func (bk *BoxKernel) Value(x []int) float64 {
	for i := range x {
		if x[i] > bk.Width || x[i] < -bk.Width {
			return 0.0
		}
	}
	return 1.0
}

// Blur applies a blurring kernel to the data. domainSize specifies the size of the
// domain in each direction. Thus, if domain size is []int{5, 5}, the length of data
// must be 25 and it is assumed that it represents a 5x5 domain. kernel is a blurring
// kernel.
func Blur(data MutableSlice, domainSize []int, kernel BlurKernel) {
	end := make([]int, len(domainSize))
	shift := make([]int, len(domainSize))
	shiftedPos := make([]int, len(domainSize))

	for i := range end {
		end[i] = 2*kernel.Cutoff() + 1
	}

	dataCpy := make([]float64, data.Len())
	for i := range dataCpy {
		dataCpy[i] = data.Get(i)
	}

	for i := range dataCpy {
		prod := NewProduct(end)
		center := Pos(domainSize, i)
		newValue := 0.0
		kernelIntegral := 0.0
		for pos := prod.Next(); pos != nil; pos = prod.Next() {
			for j := range shiftedPos {
				shift[j] = pos[j] - kernel.Cutoff()
				shiftedPos[j] = center[j] + shift[j]
			}
			Wrap(shiftedPos, domainSize)
			newValue += dataCpy[NodeIdx(domainSize, shiftedPos)] * kernel.Value(shift)
			kernelIntegral += kernel.Value(shift)
		}
		data.Set(i, newValue/kernelIntegral)
	}
}
