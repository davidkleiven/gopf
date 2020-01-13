package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gosfft/sfft"
)

func FreqVolConserved(i int) []float64 {
	return make([]float64, 2)
}

func TestVolConserve(t *testing.T) {
	N := 16
	dt := 0.01
	vol := NewVolumeConservingLP("myfield", "indicator", dt, N*N)

	indicator := NewField("indicator", N*N, nil)
	myfield := NewField("myfield", N*N, nil)

	for i := range indicator.Data {
		if i > len(indicator.Data)/2 {
			indicator.Data[i] = complex(1.0, 0.0)
		}
		myfield.Data[i] = complex(1.0, 0.0)
	}

	bricks := make(map[string]Brick)
	bricks["indicator"] = &indicator
	bricks["myfield"] = &myfield

	ft := sfft.NewFFT2(N, N)
	ft.FFT(indicator.Data)

	function := vol.Construct(bricks)
	rate := 0.01
	correction := make([]complex128, N*N)

	// Let field change accordint to dmyfield/dt = rate, and confirm that the volume conservation conserves
	// the volume
	for i := 0; i < 100; i++ {
		ft.IFFT(indicator.Data)
		DivRealScalar(indicator.Data, float64(len(indicator.Data)))
		correction = function(FreqVolConserved, 0.0, correction)
		for j := range myfield.Data {
			myfield.Data[j] += complex(dt, 0.0) * (complex(rate, 0.0) + correction[j])
		}
		ft.FFT(indicator.Data)
		vol.OnStepFinished(0.0, bricks)
	}

	volume := 0.0
	for j := range myfield.Data {
		volume += real(myfield.Data[j])
	}
	expectChange := 2.0 * dt * float64(N*N) * rate

	tol := 1e-10
	if math.Abs(volume-expectChange-256.0) > tol {
		t.Errorf("Volume conserving constraint does not keep volume conserved. After %f before 256", volume-expectChange)
	}
}
