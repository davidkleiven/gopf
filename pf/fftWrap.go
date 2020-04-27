package pf

import "github.com/barnex/fftw"

// FFTWWrapper implemente the FourierTransform interface
type FFTWWrapper struct {
	PlanFFT    fftw.Plan
	PlanIFFT   fftw.Plan
	Data       []complex128
	Dimensions []int
}

// NewFFTW returns a new FFTWWrapper
func NewFFTW(n []int) *FFTWWrapper {
	var transform FFTWWrapper
	transform.Data = make([]complex128, ProdInt(n))
	transform.PlanFFT = fftw.PlanZ2Z(n, transform.Data, transform.Data, -1, fftw.MEASURE)
	transform.PlanIFFT = fftw.PlanZ2Z(n, transform.Data, transform.Data, 1, fftw.MEASURE)
	transform.Dimensions = n
	return &transform
}

// FFT performs forward fourier transform
func (fw *FFTWWrapper) FFT(data []complex128) []complex128 {
	copy(fw.Data, data)
	fw.PlanFFT.Execute()
	copy(data, fw.Data)
	return data
}

// IFFT performs inferse fourier transform
func (fw *FFTWWrapper) IFFT(data []complex128) []complex128 {
	copy(fw.Data, data)
	fw.PlanIFFT.Execute()
	copy(data, fw.Data)
	return data
}

// col returns the column corresponding to node number i
func (fw *FFTWWrapper) col(i int) int {
	return i % fw.Dimensions[1]
}

// row returns the row corresponding to node number i
func (fw *FFTWWrapper) row(i int) int {
	return (i / fw.Dimensions[1]) % fw.Dimensions[0]
}

// depth returns the depth corresponding to node number i
func (fw *FFTWWrapper) depth(i int) int {
	return i / (fw.Dimensions[0] * fw.Dimensions[1])
}

// Freq returns the frequency corresponding to site i
func (fw *FFTWWrapper) Freq(i int) []float64 {
	res := make([]float64, len(fw.Dimensions))
	c := fw.col(i)
	r := fw.row(i)
	res[1] = float64(c) / float64(fw.Dimensions[1])
	res[0] = float64(r) / float64(fw.Dimensions[0])

	if len(res) > 2 {
		d := fw.depth(i)
		res[2] = float64(d) / float64(fw.Dimensions[2])
	}
	for i := range res {
		if res[i] > 0.5 {
			res[i] -= 1.0
		}
	}
	return res
}

// ConjugateNode returns the node that corresponds to the negative frequency
// of the node being passed
func (fw *FFTWWrapper) ConjugateNode(i int) int {
	c := fw.col(i)
	r := fw.row(i)

	conjC := (fw.Dimensions[1] - c) % fw.Dimensions[1]
	conjR := (fw.Dimensions[0] - r) % fw.Dimensions[0]
	conjD := 0

	nr := fw.Dimensions[0]
	nc := fw.Dimensions[1]

	if len(fw.Dimensions) == 3 {
		d := fw.depth(i)
		conjD = (fw.Dimensions[2] - d) % fw.Dimensions[2]
	}

	return conjD*nr*nc + conjR*nc + conjC
}
