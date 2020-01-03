package pf

// A PointMonitor is a monitor that monitors the solution at a given pixel/voxel
type PointMonitor struct {
	Data  []float64
	Site  int
	Field string
}

// NewPointMonitor returns a new instance of PointMonitor
func NewPointMonitor(site int, field string) PointMonitor {
	return PointMonitor{
		Data:  []float64{},
		Site:  site,
		Field: field,
	}
}

// Add adds a new value to the monitor
func (p *PointMonitor) Add(value float64) {
	p.Data = append(p.Data, value)
}
