package pfutil

import "testing"

func TestNode2PosRoundTrip(t *testing.T) {
	for i, test := range []struct {
		nodeNum    int
		domainSize []int
	}{
		{
			nodeNum:    16,
			domainSize: []int{5, 7},
		},
		{
			nodeNum:    124,
			domainSize: []int{11, 12, 13},
		},
	} {
		idx := Pos(test.domainSize, test.nodeNum)
		n := NodeIdx(test.domainSize, idx)
		if n != test.nodeNum {
			t.Errorf("Test #%d: Expected %d got %d", i, test.nodeNum, idx)
		}
	}
}
