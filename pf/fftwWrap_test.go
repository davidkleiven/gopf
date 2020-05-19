package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
	"github.com/davidkleiven/gosfft/sfft"
	"gonum.org/v1/gonum/floats"
)

func TestFFTWWrapConsistency(t *testing.T) {
	nx := 8
	ny := 16
	data := make([]complex128, nx*ny)
	for i := range data {
		data[i] = complex(float64(i), 0.0)
	}

	ft := sfft.NewFFT2(nx, ny)
	ftWrap := NewFFTW([]int{nx, ny})

	// Check that the frequency function match
	for i := range data {
		f1 := ft.Freq(i)
		f2 := ftWrap.Freq(i)
		if !floats.EqualApprox(f1, f2, 1e-8) {
			t.Errorf("Expected %v got %v\n", f1, f2)
		}
	}

	dataCpy := make([]complex128, len(data))
	copy(dataCpy, data)
	ft.FFT(data)
	ftWrap.FFT(dataCpy)
	tol := 1e-6
	for i := range data {
		re1 := real(data[i])
		im1 := imag(data[i])
		re2 := real(dataCpy[i])
		im2 := imag(dataCpy[i])
		if math.Abs(re1-re2) > tol || math.Abs(im1-im2) > tol {
			t.Errorf("Expected (%f, %f) got (%f, %f)\n", re1, im1, re2, im2)
		}
	}

	ft.IFFT(data)
	ftWrap.IFFT(dataCpy)
	for i := range data {
		re1 := real(data[i])
		im1 := imag(data[i])
		re2 := real(dataCpy[i])
		im2 := imag(dataCpy[i])
		if math.Abs(re1-re2) > tol || math.Abs(im1-im2) > tol {
			t.Errorf("Expected (%f, %f) got (%f, %f)\n", re1, im1, re2, im2)
		}
	}
}

func TestConjugateNode(t *testing.T) {
	for i, test := range []struct {
		Dim []int
	}{
		{
			Dim: []int{8, 8},
		},
		{
			Dim: []int{9, 9},
		},
		{
			Dim: []int{8, 8, 8},
		},
		{
			Dim: []int{9, 9, 9},
		},
	} {
		ft := NewFFTW(test.Dim)
		for j := 0; j < pfutil.ProdInt(test.Dim); j++ {
			f1 := ft.Freq(j)
			nodeIdx := ft.ConjugateNode(j)
			f2 := ft.Freq(nodeIdx)

			tol := 1e-10
			for k := range f1 {
				if math.Abs(f1[k]+f2[k]) > tol && math.Abs(f1[k]) < 0.5-tol {
					t.Errorf("Test #%d: %v and %v are not conjugate pairs\n", i, f1, f2)
				}
			}
		}
	}
}
