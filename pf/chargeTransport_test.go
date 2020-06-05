package pf

import (
	"math"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
)

func TestChargeTransport(t *testing.T) {
	N := 32
	for i, test := range []struct {
		Conductivity     func(i int) []float64
		ExternalField    []float64
		ChargeDensity    func(i int) float64
		ExpectDivCurrent func(i int) float64
		ExpectedCurrent  func(i int) []float64
		Tol              float64
	}{
		// Test 0: Zero charge density, isotropic conductivity
		{
			Conductivity: func(i int) []float64 {
				return []float64{1.0, 1.0, 0.0}
			},
			ExternalField: []float64{1.0, 0.0},
			ChargeDensity: func(i int) float64 {
				return 0.0
			},
			ExpectDivCurrent: func(i int) float64 {
				return 0.0
			},
			ExpectedCurrent: func(i int) []float64 {
				return []float64{1.0, 0.0}
			},
			Tol: 1e-10,
		},
		// Test 1: Sinosoidal charge density
		{
			Conductivity: func(i int) []float64 {
				return []float64{1.0, 1.0, 0.0}
			},
			ExternalField: []float64{1.0, 0.0},
			ChargeDensity: func(i int) float64 {
				pos := pfutil.Pos([]int{N, N}, i)
				x := float64(pos[0]) / float64(N)
				return math.Sin(2.0 * math.Pi * x)
			},
			ExpectDivCurrent: func(i int) float64 {
				pos := pfutil.Pos([]int{N, N}, i)
				x := float64(pos[0]) / float64(N)
				return -math.Sin(2.0 * math.Pi * x)
			},
			ExpectedCurrent: func(i int) []float64 {
				pos := pfutil.Pos([]int{N, N}, i)
				x := float64(pos[0]) / float64(N)
				currentX := 1.0 - float64(N)*math.Cos(2.0*math.Pi*x)/(2.0*math.Pi)
				return []float64{currentX, 0.0}
			},
			Tol: 1e-10,
		},
	} {
		field := NewField("density", N*N, nil)
		ft := NewFFTW([]int{N, N})
		chargeTransp := ChargeTransport{
			Conductivity:  test.Conductivity,
			ExternalField: test.ExternalField,
			Field:         "density",
			FT:            ft,
		}

		origData := make([]complex128, len(field.Data))
		for j := range field.Data {
			field.Data[j] = complex(test.ChargeDensity(j), 0.0)
		}
		copy(origData, field.Data)

		field.Data = ft.FFT(field.Data)
		bricks := make(map[string]Brick)
		bricks["density"] = field
		function := chargeTransp.Construct(bricks)

		result := make([]complex128, N*N)
		function(ft.Freq, 0.0, result)
		ft.IFFT(result)
		pfutil.DivRealScalar(result, float64(N*N))

		for j := range result {
			re := real(result[j])
			imag := imag(result[j])
			expect := test.ExpectDivCurrent(j)
			if math.Abs(re-expect) > test.Tol || math.Abs(imag) > test.Tol {
				t.Errorf("Test #%d: Expected %f got (%f, %f)\n", i, expect, re, imag)
			}
		}

		current := chargeTransp.Current(field, len(field.Data), false)
		for j := range field.Data {
			for d := 0; d < 2; d++ {
				expect := test.ExpectedCurrent(j)[d]
				got := current[d][j]
				if math.Abs(expect-got) > test.Tol {
					t.Errorf("Test #%d: Current %d: Expected %f got %f\n", i, d, expect, got)
				}
			}
		}

		copy(field.Data, origData)
		currentReal := chargeTransp.Current(field, len(field.Data), true)
		tol := 1e-10
		for j := range current[0] {
			for comp := 0; comp < 2; comp++ {
				re1 := currentReal[comp][j]
				re2 := current[comp][j]
				if math.Abs(re1-re2) > tol {
					t.Errorf("Expected %f got %f\n", re1, re2)
				}
			}
		}
	}
}
