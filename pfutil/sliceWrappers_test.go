package pfutil

import (
	"math"
	"testing"
)

func TestWrappers(t *testing.T) {
	type TestCase struct {
		Index     int
		GetExpect float64
		SetValue  float64
		Length    int
	}

	for i, test := range []struct {
		Data MutableSlice
		Test TestCase
	}{
		{
			Data: &RealSlice{
				Data: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			},
			Test: TestCase{
				Index:     3,
				GetExpect: 4.0,
				SetValue:  -3.0,
				Length:    5,
			},
		},
		{
			Data: &RealPartSlice{
				Data: []complex128{complex(1.0, 0.0), complex(-1.0, -2.0), complex(7.0, -5.0)},
			},
			Test: TestCase{
				Index:     1,
				GetExpect: -1.0,
				SetValue:  20.0,
				Length:    3,
			},
		},
	} {

		if test.Data.Len() != test.Test.Length {
			t.Errorf("Test #%d: Expected %d got %d\n", i, test.Test.Length, test.Data.Len())
		}

		idx := test.Test.Index
		if math.Abs(test.Data.Get(idx)-test.Test.GetExpect) > 1e-10 {
			t.Errorf("Test #%d: Expected %f got %f\n", i, test.Test.GetExpect, test.Data.Get(idx))
		}

		test.Data.Set(idx, test.Test.SetValue)

		if math.Abs(test.Data.Get(idx)-test.Test.SetValue) > 1e-10 {
			t.Errorf("Test: #%d: Expected %f got %f\n", i, test.Test.SetValue, test.Data.Get(idx))
		}

	}
}
