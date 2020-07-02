package pfutil

// ImmutableSlice is a type for wrapping slices that can be accessed, but not altered slices
type ImmutableSlice interface {
	Get(i int) float64
	Len() int
}

// MutableSlice is an interface for slice wrapeprs where the underlying data array can be
// both accessed and altered
type MutableSlice interface {
	ImmutableSlice
	Set(i int, v float64)
}

// RealSlice implements the MutableSlice interface when the underlying data is a real array
type RealSlice struct {
	Data []float64
}

// Get returns the value at position i
func (rs *RealSlice) Get(i int) float64 {
	return rs.Data[i]
}

// Set sets the value at index i equal to v
func (rs *RealSlice) Set(i int, v float64) {
	rs.Data[i] = v
}

// Len returns the length of the underlying data array
func (rs *RealSlice) Len() int {
	return len(rs.Data)
}

// RealPartSlice implements the MutableSlice interface. It only operates on the real part of the
// underlying complex array
type RealPartSlice struct {
	Data []complex128
}

// Get returns the value at position i
func (rps *RealPartSlice) Get(i int) float64 {
	return real(rps.Data[i])
}

// Set sets the real part at position i to the new value
func (rps *RealPartSlice) Set(i int, v float64) {
	rps.Data[i] = complex(v, imag(rps.Data[i]))
}

// Len returns the length of the underlying data array
func (rps *RealPartSlice) Len() int {
	return len(rps.Data)
}
