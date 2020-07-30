// +build ignore

package main

import (
	"database/sql"
	"math"
	"math/rand"

	"github.com/davidkleiven/gopf/pf"
	"github.com/davidkleiven/gopf/pfutil"

	_ "github.com/mattn/go-sqlite3"
)

// CircleMonitor is a type that observes a field inside
type CircleMonitor struct {
	Radius int
	X      int
	Y      int
	Name   string
}

// CircleMonitorResult contains the mean and the standrad deviation of a field
type CircleMonitorResult struct {
	Mean float64
	Std  float64
}

// Eval runs the circle evaluated the field
func (cm *CircleMonitor) Eval(field pf.Brick, domainSize []int) CircleMonitorResult {
	var result CircleMonitorResult
	num := 0
	for ix := cm.X - cm.Radius; ix < cm.X+cm.Radius; ix++ {
		for iy := cm.Y - cm.Radius; iy < cm.Y+cm.Radius; iy++ {
			x := ix - cm.X
			y := iy - cm.Y
			rSq := x*x + y*y
			if rSq < cm.Radius*cm.Radius {
				idx := pfutil.NodeIdx(domainSize, []int{ix, iy})
				result.Mean += real(field.Get(idx))
				result.Std += math.Pow(real(field.Get(idx)), 2)
				num++
			}
		}
	}
	result.Mean /= float64(num)
	result.Std = math.Sqrt(result.Std/float64(num) - result.Mean*result.Mean)
	return result
}

// DBWriter is a type for writing records to the database
type DBWriter struct {
	Monitors []CircleMonitor
	DB       pf.FieldDB
	numCalls int
}

// Eval is the pf.SolverCB type and can be attached as a callback
// to the Solver
func (db *DBWriter) Eval(s *pf.Solver, timestep int) {
	data := make(map[string]float64)
	for _, monitor := range db.Monitors {
		res := monitor.Eval(s.Model.Bricks["conc"], db.DB.DomainSize)
		data[monitor.Name+"_mean"] = res.Mean
		data[monitor.Name+"_std"] = res.Std
	}
	db.DB.TimeSeries(data, timestep)

	// We don't write field data as often as we write timeseries data
	// due to size
	if timestep%10 == 0 {
		db.DB.SaveFields(s, timestep)
	}
}

func main() {
	dbName := "./diffusion.db"

	sqlDB, _ := sql.Open("sqlite3", dbName)
	fieldDB := pf.FieldDB{
		DB:         sqlDB,
		DomainSize: []int{128, 128},
	}

	comment := "This is a very long comment. We use a very long comment here "
	comment += "in order to check that the command line interface manages to split the lines correctly "
	comment += "when printing it. The comment should be printed in a way such that each line has a maximum "
	comment += "width, the reminding part of the comment should be written on the next line. Furthermore, "
	comment += "the simulation ID should only be displayed once per comment."

	fieldDB.Comment(comment)

	// Add some attributes
	attr := make(map[string]float64)
	attr["start"] = 0.1
	attr["meanConc"] = 0.5
	fieldDB.SetAttr(attr)

	// Add some text attributes
	attrTxt := make(map[string]string)
	attrTxt["txt"] = "textattr"
	fieldDB.SetTextAttr(attrTxt)

	// Create a model and add some fields
	conc := pf.NewField("conc", 128*128, nil)
	for i := range conc.Data {
		conc.Data[i] = complex(rand.Float64(), 0.0)
	}

	model := pf.NewModel()
	model.AddField(conc)
	model.AddEquation("dconc/dt = LAP conc")
	solver := pf.NewSolver(&model, fieldDB.DomainSize, 0.1)

	// Let's run the calculation. We want to track the various time series data.
	// We construct a few examples
	// 1. Mean concentration inside a circle of radius 5 located at (54, 54)
	// 2. Mean concenration inside a circle of radius 5 located at (72, 72)
	// 3. Variation of the concentration inside a circle of radius 5 located at (54, 54)
	// 4. Variation of the concentration inside a circle of radius 5 located at (54, 54)

	writer := DBWriter{
		DB: fieldDB,
		Monitors: []CircleMonitor{
			{
				Name:   "monitor5454",
				X:      54,
				Y:      54,
				Radius: 5,
			},
			{
				Name:   "monitor7272",
				X:      72,
				Y:      72,
				Radius: 5,
			},
		},
	}

	// Add the evaluate function
	solver.AddCallback(writer.Eval)

	// Solve the system
	solver.Solve(20, 10)
}
