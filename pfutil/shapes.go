package pfutil

import "fmt"

// Shape is represents a generic interface to represent geometrical shapes
type Shape interface {
	// InteriorPoint returns true if the passed point is inside the shape
	InteriorPoint(pos []float64) bool

	// BBox returns the bounding box of the shape
	BBox() BoundingBox
}

// BoundingBox represents the bounding box for the shape
type BoundingBox struct {
	Min []int
	Max []int
}

func (b *BoundingBox) check() {
	if len(b.Min) != 3 || len(b.Max) != 3 {
		panic("BoundinBox: Both Min and Max must have length 3")
	}
	for i := range b.Min {
		if b.Min[i] > b.Max[i] {
			panic("BoundingBox: Min is larger than Max")
		}
	}
}

// Box represents a rectangle in 2D and a box in 3D
type Box struct {
	Diagonal []float64
}

// InteriorPoint returns true if the passed point is inside the box
func (b *Box) InteriorPoint(pos []float64) bool {
	if len(pos) > len(b.Diagonal) {
		panic(fmt.Sprintf("Box: Passed position has length %d. Box has dimension %d\n", len(pos), len(b.Diagonal)))
	}

	for i := range pos {
		if pos[i] > b.Diagonal[i]/2.0 || pos[i] < -b.Diagonal[i]/2.0 {
			return false
		}
	}
	return true
}

// BBox returns the bounding box of the shape
func (b *Box) BBox() BoundingBox {
	bbox := BoundingBox{
		Min: []int{0, 0, 0},
		Max: make([]int, 3),
	}

	for i := range b.Diagonal {
		bbox.Min[i] = -int(b.Diagonal[i]/2.0) - 1
		bbox.Max[i] = int(b.Diagonal[i]/2.0) + 1
	}
	return bbox
}

// Ball represents a ball in n dimensions. In 2D this is a circle and
// in 3D this is a sphere
type Ball struct {
	Radius float64
}

// InteriorPoint returns true if the passed point is inside the ball
func (b *Ball) InteriorPoint(pos []float64) bool {
	rSq := Dot(pos, pos)
	if rSq < b.Radius*b.Radius {
		return true
	}
	return false
}

// BBox returns the bounding box
func (b *Ball) BBox() BoundingBox {
	rmin := -int(b.Radius) - 1
	rmax := int(b.Radius) + 1
	bbox := BoundingBox{
		Min: make([]int, 3),
		Max: make([]int, 3),
	}
	for i := 0; i < 3; i++ {
		bbox.Min[i] = rmin
		bbox.Max[i] = rmax
	}
	return bbox
}

// Draw draws the shape onto the passed grid. The passed transfformations is applied
// to the shape prior to drawing. If a point is inside the passed shape, the value of
// that grid point will be set to value
func Draw(shape Shape, grid *Grid, transformation *Affine, value float64) {
	if transformation == nil {
		identity := Identity()
		transformation = &identity
	}

	bbox := shape.BBox()
	bbox.check()
	dim := len(grid.Dims)

	for ix := bbox.Min[0]; ix <= bbox.Max[0]; ix++ {
		for iy := bbox.Min[1]; iy <= bbox.Max[1]; iy++ {
			for iz := bbox.Min[2]; iz <= bbox.Max[2]; iz++ {
				pos := []float64{float64(ix), float64(iy), float64(iz)}
				if shape.InteriorPoint(pos[:dim]) {
					transformation.Apply(pos)
					intPos := []int{int(pos[0]), int(pos[1]), int(pos[2])}[:dim]
					Wrap(intPos, grid.Dims)
					grid.Data[grid.Index(intPos)] = value
				}
			}
		}
	}
}
