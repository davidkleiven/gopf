package pfutil

// Product implements a generic product. It can be used to loop over
// all combinations
// Example:
// The following is equivalent to a double for loo from 0 to 3
//
// prod := NewProduct([]{3, 3})
// for idx := prod.Current; prod.Next() != nil; idx = prod.Current {
// 		Do something
// }
type Product struct {
	End     []int
	Current []int
	isFirst bool
}

// NewProduct returns a new product iterator.
func NewProduct(end []int) Product {
	return Product{
		End:     end,
		Current: make([]int, len(end)),
		isFirst: true,
	}
}

// Next returns the next integer set
func (p *Product) Next() []int {
	if p.isFirst {
		p.isFirst = false
		return p.Current
	}
	p.Current[len(p.Current)-1]++

	if p.Current[len(p.Current)-1] == p.End[len(p.End)-1] {
		p.Current[len(p.Current)-1] = 0
		for i := len(p.Current) - 2; i >= 0; i-- {
			p.Current[i]++
			if p.Current[i] < p.End[i] {
				return p.Current
			}
			p.Current[i] = 0
		}
	}

	for i := range p.Current {
		if p.Current[i] != 0 {
			return p.Current
		}
	}
	return nil
}
