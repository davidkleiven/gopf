package pfc

import "testing"

func TestNumEquiv(t *testing.T) {
	for i, test := range []struct {
		Miller Miller
		expect int
	}{
		{
			Miller: Miller{H: 1, K: 1, L: 1},
			expect: 8,
		},
		{
			Miller: Miller{H: 2, K: 0, L: 0},
			expect: 6,
		},
		{
			Miller: Miller{H: 1, K: -1, L: 1},
			expect: 8,
		},
		{
			Miller: Miller{H: 1, K: 3, L: 0},
			expect: 24,
		},
		{
			Miller: Miller{H: 0, K: 0, L: 0},
			expect: 1,
		},
	} {
		num := NumEquivalent(test.Miller)
		if num != test.expect {
			t.Errorf("Test #%d: Expected %d got %d\n", i, test.expect, num)
		}
	}
}

func TestEquivalentMillerNoPermutations(t *testing.T) {
	for i, test := range []struct {
		Miller Miller
		Expect []Miller
	}{
		{
			Miller: Miller{1, 0, 0},
			Expect: []Miller{
				Miller{-1, 0, 0},
				Miller{1, 0, 0},
			},
		},
		{
			Miller: Miller{1, 2, 0},
			Expect: []Miller{
				Miller{-1, -2, 0},
				Miller{-1, 2, 0},
				Miller{1, -2, 0},
				Miller{1, 2, 0},
			},
		},
		{
			Miller: Miller{0, 2, 0},
			Expect: []Miller{
				Miller{0, -2, 0},
				Miller{0, 2, 0},
			},
		},
		{
			Miller: Miller{0, 0, 2},
			Expect: []Miller{
				Miller{0, 0, -2},
				Miller{0, 0, 2},
			},
		},
	} {
		res := EquivalentMillerNoPermutations(test.Miller)
		if len(res) != len(test.Expect) {
			t.Errorf("Test #%d: Expected %d equivalent planes. Got %d\n", i, len(test.Expect), len(res))
		}

		for j := range res {
			if !res[j].Equal(test.Expect[j]) {
				t.Errorf("Test #%d: Expected %v. Got %v\n", i, res[j], test.Expect[j])
			}
		}
	}
}

func TestEquivalentMiller(t *testing.T) {
	for i, test := range []struct {
		Miller Miller
		Expect []Miller
	}{
		{
			Miller: Miller{1, 0, 0},
			Expect: []Miller{
				Miller{-1, 0, 0},
				Miller{1, 0, 0},
				Miller{0, -1, 0},
				Miller{0, 1, 0},
				Miller{0, 0, -1},
				Miller{0, 0, 1},
			},
		},
		{
			Miller: Miller{1, 1, 1},
			Expect: []Miller{
				Miller{1, 1, 1},
				Miller{-1, 1, 1},
				Miller{1, -1, 1},
				Miller{1, 1, -1},
				Miller{-1, 1, -1},
				Miller{1, -1, -1},
				Miller{-1, -1, -1},
				Miller{-1, -1, 1},
			},
		},
		{
			Miller: Miller{2, 1, 3},
			Expect: []Miller{
				Miller{2, 1, 3},
				Miller{1, 2, 3},
				Miller{1, 3, 2},
				Miller{3, 1, 2},
				Miller{3, 2, 1},
				Miller{2, 3, 1},

				Miller{-2, 1, 3},
				Miller{1, -2, 3},
				Miller{1, 3, -2},
				Miller{3, 1, -2},
				Miller{3, -2, 1},
				Miller{-2, 3, 1},

				Miller{2, -1, 3},
				Miller{-1, 2, 3},
				Miller{-1, 3, 2},
				Miller{3, -1, 2},
				Miller{3, 2, -1},
				Miller{2, 3, -1},

				Miller{2, 1, -3},
				Miller{1, 2, -3},
				Miller{1, -3, 2},
				Miller{-3, 1, 2},
				Miller{-3, 2, 1},
				Miller{2, -3, 1},

				Miller{-2, -1, 3},
				Miller{-1, -2, 3},
				Miller{-1, 3, -2},
				Miller{3, -1, -2},
				Miller{3, -2, -1},
				Miller{-2, 3, -1},

				Miller{2, -1, -3},
				Miller{-1, 2, -3},
				Miller{-1, -3, 2},
				Miller{-3, -1, 2},
				Miller{-3, 2, -1},
				Miller{2, -3, -1},

				Miller{-2, 1, -3},
				Miller{1, -2, -3},
				Miller{1, -3, -2},
				Miller{-3, 1, -2},
				Miller{-3, -2, 1},
				Miller{-2, -3, 1},

				Miller{-2, -1, -3},
				Miller{-1, -2, -3},
				Miller{-1, -3, -2},
				Miller{-3, -1, -2},
				Miller{-3, -2, -1},
				Miller{-2, -3, -1},
			},
		},
	} {
		res := EquivalentMiller(test.Miller)
		if len(res) != len(test.Expect) {
			t.Errorf("Test #%d: Expected %d equivalent planes. Got %d.\nAll indices\n:%v\n", i, len(test.Expect), len(res), res)
			return
		}

		found := make([]bool, len(test.Expect))
	allMCheck:
		for j := range res {
			for k := range test.Expect {
				if res[j].Equal(test.Expect[k]) {
					if found[j] {
						t.Errorf("%v exists more than one time\nAll indices:%v\n", res[j], res)
						break allMCheck
					} else {
						found[j] = true
					}
				}
			}
		}

		for j := range found {
			if !found[j] {
				t.Errorf("%v does note exist among the miller indices. All indices\n%v\n\n", res[j], res)
				break
			}
		}
	}
}
