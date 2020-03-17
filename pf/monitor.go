package pf

// Monitor is a generic interface monitor structures
type Monitor interface {
	Add(bricks map[string]Brick)
}

// A PointMonitor is a monitor that monitors the solution at a given pixel/voxel
type PointMonitor struct {
	Data  []float64
	Site  int
	Field string
	Name  string
}

// NewPointMonitor returns a new instance of PointMonitor
func NewPointMonitor(site int, field string) PointMonitor {
	return PointMonitor{
		Data:  []float64{},
		Site:  site,
		Field: field,
		Name:  "PointMonitor",
	}
}

// Add adds a new value to the monitor
func (p *PointMonitor) Add(bricks map[string]Brick) {
	value := real(bricks[p.Field].Get(p.Site))
	p.Data = append(p.Data, value)
}
