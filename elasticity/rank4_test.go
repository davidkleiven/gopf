package elasticity

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

func TestIsotropic(t *testing.T) {
	B := 61.4
	poisson := 0.3
	tensor := Isotropic(B, poisson)
	data := make([]float64, 81)
	copy(data, tensor.Data)
	for i, test := range []struct {
		axis  int
		angle float64
	}{
		{
			axis:  0,
			angle: 14.0 * math.Pi / 180.0,
		},
		{
			axis:  1,
			angle: 56.0 * math.Pi / 180.0,
		},
		{
			axis:  2,
			angle: -56.0 * math.Pi / 180.0,
		},
	} {
		rot := RotationMatrix(test.angle, test.axis)
		tensor.Rotate(rot)
		if !floats.EqualApprox(tensor.Data, data, 1e-10) {
			t.Errorf("Test #%d: Isotropic tensor is not invariant under rotations", i)
		}
	}
}

func TestCubicMaterial(t *testing.T) {
	C11 := 110.0
	C12 := 60.0
	C44 := 30.0
	matprop := CubicMaterial(C11, C12, C44)
	data := make([]float64, 81)
	copy(data, matprop.Data)

	for i, test := range []struct {
		rot         *mat.Dense
		expectEqual bool
	}{
		{
			rot:         mat.NewDense(3, 3, []float64{1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, 1.0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{-1.0, 0.0, 0.0, 0.0, -1.0, 0.0, 0.0, 0.0, 1.0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{-1.0, 0.0, 0.0, 0.0, 1.0, 0.0, 0.0, 0.0, -1.0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{1.0, 0.0, 0.0, 0.0, -1.0, 0.0, 0.0, 0.0, -1.0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, 1, 1, 0, 0, 0, 1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, 1, -1, 0, 0, 0, -1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, -1, -1, 0, 0, 0, 1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, -1, 1, 0, 0, 0, -1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 1, 0, 0, 0, 1, 1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, -1, 0, 0, 0, 1, -1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 1, 0, 0, 0, -1, -1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, -1, 0, 0, 0, -1, 1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 1, 0, 1, 0, 0, 0, 0, -1}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, -1, 0, -1, 0, 0, 0, 0, -1}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 1, 0, -1, 0, 0, 0, 0, 1}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, -1, 0, 1, 0, 0, 0, 0, 1}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{1, 0, 0, 0, 0, 1, 0, -1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{-1, 0, 0, 0, 0, 1, 0, 1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{-1, 0, 0, 0, 0, -1, 0, -1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{1, 0, 0, 0, 0, -1, 0, 1, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, 1, 0, 1, 0, -1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, 1, 0, -1, 0, 1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, -1, 0, 1, 0, 1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         mat.NewDense(3, 3, []float64{0, 0, -1, 0, -1, 0, -1, 0, 0}),
			expectEqual: true,
		},
		{
			rot:         RotationMatrix(43.0, 0),
			expectEqual: false,
		},
	} {
		matprop.Rotate(test.rot)

		if test.expectEqual {
			if !floats.EqualApprox(matprop.Data, data, 1e-10) {
				t.Errorf("Test #%d:\nExpected\n%v\ngot\n%v\n", i, data, matprop.Data)
			}
		} else {
			if floats.EqualApprox(matprop.Data, data, 1e-10) {
				t.Errorf("Test #%d: Tensor invariant, although it should not be...", i)
			}
		}
	}
}

func TestFromVoigt(t *testing.T) {
	strainEng := []float64{1.0, 2.0, 3.0, 8.0, 10.0, 12.0}
	strain := mat.NewDense(3, 3, []float64{1.0, 6.0, 5.0,
		6.0, 2.0, 4.0,
		5.0, 4.0, 3.0})

	tensor := make([]float64, 36)
	for i := range tensor {
		tensor[i] = float64(i)
	}
	res := FromFlatVoigt(tensor)
	voigtTensor := mat.NewDense(6, 6, tensor)

	// Check that stresses are the same
	stress := res.ContractLast(strain)
	stressVoigt := mat.NewVecDense(6, nil)
	stressVoigt.MulVec(voigtTensor, mat.NewVecDense(6, strainEng))
	tol := 1e-10
	for i := 0; i < 3; i++ {
		for j := i; j < 3; j++ {
			v1 := stress.At(i, j)
			v2 := 0.0
			if i == j {
				v2 = stressVoigt.AtVec(i)
			} else if i == 0 && j == 1 {
				v2 = stressVoigt.AtVec(5)
			} else if i == 0 && j == 2 {
				v2 = stressVoigt.AtVec(4)
			} else if i == 1 && j == 2 {
				v2 = stressVoigt.AtVec(3)
			}

			if math.Abs(v1-v2) > tol {
				t.Errorf("At (%d, %d): Expected %f got %f\n", i, j, v1, v2)
			}
		}
	}
}
