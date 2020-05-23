package pf

import "testing"

const (
	CONSTANT = iota
	FIELD1
	FIELD2
	FIELD1XFIELD2
	LAPLACIAN
	NODEIDX
)

// Define a set of terms
type testterm struct {
	desc int
}

func (tt *testterm) Construct(bricks map[string]Brick) Term {
	return func(freq Frequency, t float64, field []complex128) {
		for i := range field {
			switch tt.desc {
			case CONSTANT:
				field[i] = complex(1.0, 0.0)
			case FIELD1:
				field[i] = bricks["field1"].Get(i)
			case FIELD2:
				field[i] = bricks["field2"].Get(i)
			case FIELD1XFIELD2:
				field[i] = bricks["field2"].Get(i)
			case LAPLACIAN:
				lap := LaplacianN{Power: 1}
				field[i] = complex(1.0, 0.0)
				field = lap.Eval(freq, field)
			case NODEIDX:
				field[i] = complex(float64(i), 0.0)
			}
		}
	}
}

func (tt *testterm) OnStepFinished(t float64, bricks map[string]Brick) {}

func TestIsImplicit(t *testing.T) {
	bricks := make(map[string]Brick)
	bricks["field1"] = NewField("field1", 4, nil)
	bricks["field2"] = NewField("field2", 4, nil)

	freq := func(i int) []float64 { return []float64{0.2, 0.2} }
	for i, test := range []struct {
		desc int
		want bool
	}{
		{
			desc: CONSTANT,
			want: true,
		},
		{
			desc: FIELD1,
			want: false,
		},
		{
			desc: FIELD2,
			want: false,
		},
		{
			desc: FIELD1XFIELD2,
			want: false,
		},
		{
			desc: LAPLACIAN,
			want: true,
		},
		{
			desc: NODEIDX,
			want: true,
		},
	} {
		term := testterm{
			desc: test.desc,
		}
		if isImplicit(&term, bricks, 4, freq) != test.want {
			t.Errorf("Test #%d: Want %v got %v\n", i, test.want, !test.want)
		}
	}
}
