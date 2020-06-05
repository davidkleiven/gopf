package pf

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
	_ "github.com/mattn/go-sqlite3"
)

func TestInitialize(t *testing.T) {
	dbName := "./testinit.db"
	sqlDB, err := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	if err != nil {
		t.Errorf(fmt.Sprintf("%v\n", err))
		return
	}

	db := FieldDB{
		DB: sqlDB,
	}

	db.initialize()
	if !db.initialized {
		t.Errorf("Initialized flag not set")
	}

	if db.simID == 0 {
		t.Errorf("SimID not set")
	}

	expectTables := []string{
		"comments", "fields", "positions", "simAttributes", "simIDs",
		"simTextAttributes", "timeseries",
	}

	// Extract all table names
	rows, _ := db.DB.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	tableNames := []string{}
	var tabName string
	for rows.Next() {
		rows.Scan(&tabName)
		tableNames = append(tableNames, tabName)
	}

	if len(tableNames) != len(expectTables) {
		t.Errorf("Expected\n%v\nGot\n%v\n", expectTables, tableNames)
	}

	for i := range expectTables {
		if tableNames[i] != expectTables[i] {
			t.Errorf("Expected\n%v\nGot\n%v\n", expectTables, tableNames)
			break
		}
	}

	if db.positionTableIsPopulated() {
		t.Errorf("Position table is not populated")
	}
}

func TestPopulatePositionTable(t *testing.T) {
	dbName := "./testpositiontable.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)

	db := FieldDB{
		DB:         sqlDB,
		DomainSize: []int{3, 3},
	}
	db.initialize()
	db.populatePositionsTable()

	expect := [][]int{
		{0, 0, 0},
		{0, 1, 0},
		{0, 2, 0},
		{1, 0, 0},
		{1, 1, 0},
		{1, 2, 0},
		{2, 0, 0},
		{2, 1, 0},
		{2, 2, 0},
	}

	rows, _ := db.DB.Query("SELECT X, Y, Z FROM positions ORDER BY id")

	var x, y, z int
	count := 0
	for rows.Next() {
		rows.Scan(&x, &y, &z)
		if x != expect[count][0] || y != expect[count][1] || z != expect[count][2] {
			t.Errorf("Expected %v\nGot %d %d %d", expect[count], x, y, z)
		}
		count++
	}
}

func TestInsertRealPart(t *testing.T) {
	dbName := "./testInsertReal.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	db := FieldDB{
		DB:         sqlDB,
		DomainSize: []int{3, 3},
	}

	data := make([]complex128, 9)
	for i := range data {
		data[i] = complex(float64(i), 0.0)
	}
	db.initialize()
	db.populatePositionsTable()
	db.insertRealPart("myfield", 2, data)

	rows, _ := db.DB.Query("SELECT name, timestep, value, simID FROM fields")
	var name string
	var timestep int
	var simID int32
	var value float64
	rowCount := 0
	for rows.Next() {
		rows.Scan(&name, &timestep, &value, &simID)
		expect := float64(rowCount)
		if name != "myfield" || timestep != 2 || math.Abs(value-expect) > 1e-10 || simID != db.simID {
			t.Errorf("Expected myfield, 2, %f, %d\nGot %s, %d, %f, %d\n",
				expect, db.simID, name, timestep, value, simID)
		}
		rowCount++
	}
}

func TestSaveFields(t *testing.T) {
	field1 := NewField("field1", 9, nil)
	field2 := NewField("field2", 9, nil)
	model := NewModel()
	model.AddField(field1)
	model.AddField(field2)
	model.AddEquation("dfield1/dt = -field1")
	model.AddEquation("dfield2/dt = -field2")

	ds := []int{3, 3}
	solver := NewSolver(&model, ds, 0.1)

	dbName := "./testSaveFields.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	db := FieldDB{
		DB:         sqlDB,
		DomainSize: ds,
	}
	db.SaveFields(solver, 1)

	rows, _ := db.DB.Query("SELECT COUNT(*) FROM positions")
	var numRows int
	for rows.Next() {
		rows.Scan(&numRows)
	}
	if numRows != 9 {
		t.Errorf("Expected 9 positions got %d\n", numRows)
	}

	rows, _ = db.DB.Query("SELECT COUNT(*) FROM fields")
	for rows.Next() {
		rows.Scan(&numRows)
	}
	if numRows != 18 {
		t.Errorf("Expected 18 rows in the fields table got %d\n", numRows)
	}
}

func TestComment(t *testing.T) {
	dbName := "./testcomment.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	db := FieldDB{
		DB: sqlDB,
	}
	comment := "This is a comment"
	db.Comment(comment)
	rows, _ := db.DB.Query("SELECT simID, value FROM comments")
	var simID int32
	var value string
	for rows.Next() {
		rows.Scan(&simID, &value)
	}

	if simID != db.simID {
		t.Errorf("Expected %d got %d\n", db.simID, simID)
	}

	if value != comment {
		t.Errorf("Expected+n%s\nGot\n%s\n", comment, value)
	}
}

func TestSetAttr(t *testing.T) {
	dbName := "./testattribute.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)

	db := FieldDB{
		DB: sqlDB,
	}

	kvp := make(map[string]float64)
	kvp["temperature"] = 20.0
	kvp["concentration"] = 0.8
	db.SetAttr(kvp)

	rows, _ := db.DB.Query("SELECT key,value FROM simAttributes ORDER BY key")
	expect := []struct {
		key   string
		value float64
	}{
		{
			key:   "concentration",
			value: 0.8,
		},
		{
			key:   "temperature",
			value: 20.0,
		},
	}

	count := 0
	var key string
	var value float64
	for rows.Next() {
		rows.Scan(&key, &value)

		item := expect[count]
		count++
		if key != item.key || math.Abs(item.value-value) > 1e-10 {
			t.Errorf("Expected (%s, %f) got (%s, %f)\n", item.key, item.value, key, value)
		}
	}
}

func TestSetTextAttr(t *testing.T) {
	dbName := "./testsettextattr.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	db := FieldDB{
		DB: sqlDB,
	}

	attr := make(map[string]string)
	attr["name"] = "simulation1"
	attr["node"] = "sophus"

	db.SetTextAttr(attr)

	rows, _ := db.DB.Query("SELECT key, value FROM simTextAttributes ORDER BY key")

	expect := []struct {
		key   string
		value string
	}{
		{
			key:   "name",
			value: "simulation1",
		},
		{
			key:   "node",
			value: "sophus",
		},
	}

	count := 0
	var key, value string
	for rows.Next() {
		rows.Scan(&key, &value)

		item := expect[count]
		count++
		if key != item.key || value != item.value {
			t.Errorf("Expected (%s, %s) got (%s, %s)\n", item.key, item.value, key, value)
		}
	}
}

func TestTimeSeries(t *testing.T) {
	dbName := "./testtimeseries.db"
	sqlDB, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)
	db := FieldDB{
		DB: sqlDB,
	}

	tData := make(map[string]float64)
	tData["energy"] = -0.1
	tData["concentration"] = 0.5
	db.TimeSeries(tData, 0)
	db.TimeSeries(tData, 1)

	expect := []struct {
		key   string
		value float64
		time  int
	}{
		{
			key:   "energy",
			value: -0.1,
			time:  0,
		},
		{
			key:   "concentration",
			value: 0.5,
			time:  0,
		},
		{
			key:   "energy",
			value: -0.1,
			time:  1,
		},
		{
			key:   "concentration",
			value: 0.5,
			time:  1,
		},
	}

	// Count the number of rows
	rows, _ := db.DB.Query("SELECT COUNT(*) FROM timeseries")
	var numRows int
	for rows.Next() {
		rows.Scan(&numRows)
	}

	if numRows != len(expect) {
		t.Errorf("Expected %d rows. Got %d\n", len(expect), numRows)
		return
	}

	rows, _ = db.DB.Query("SELECT key,value,timestep FROM timeseries ORDER BY timestep ASC, key DESC")
	count := 0
	var key string
	var value float64
	var timestep int
	for rows.Next() {
		item := expect[count]
		count++
		rows.Scan(&key, &value, &timestep)
		if item.key != key || timestep != item.time || math.Abs(item.value-value) > 1e-10 {
			t.Errorf("Expected (%s, %f, %d) got (%s, %f, %d)\n", item.key, item.value, item.time,
				key, value, timestep)
		}
	}
}

type constant struct {
	Value float64
}

func (c *constant) Eval(i int, bricks map[string]Brick) complex128 {
	return complex(c.Value, 0.0)
}

func TestLoadFields(t *testing.T) {
	temperature := NewField("temperature", 16, nil)
	concentration := NewField("conc", 16, nil)
	model := NewModel()
	model.AddField(temperature)
	model.AddField(concentration)
	rateTemp := constant{
		Value: 0.1,
	}
	rateConc := constant{
		Value: -0.1,
	}

	model.RegisterFunction("RATE_TEMP", rateTemp.Eval)
	model.RegisterFunction("RATE_CONC", rateConc.Eval)

	model.AddEquation("dtemperature/dt = RATE_TEMP")
	model.AddEquation("dconcentration/dt = RATE_CONC")

	dbName := "test_load.db"
	db, _ := sql.Open("sqlite3", dbName)
	defer os.Remove(dbName)

	fieldDB := FieldDB{
		DB:         db,
		DomainSize: []int{4, 4},
	}

	solver := NewSolver(&model, fieldDB.DomainSize, 1.0)
	solver.AddCallback(fieldDB.SaveFields)
	solver.Solve(10, 1)

	// Run tests
	for step := 0; step < 10; step++ {
		fields := fieldDB.Load(int(fieldDB.simID), step)

		// Fill temperature and concentration fields with the expected values
		for i := range temperature.Data {
			temperature.Data[i] = complex(0.1*float64(step+1), 0.0)
			concentration.Data[i] = complex(-0.1*float64(step+1), 0.0)
		}
		expect := []Field{
			concentration, temperature,
		}
		for i := range fields {
			if !pfutil.CmplxEqualApprox(fields[i].Data, expect[i].Data, 1e-10) {
				t.Errorf("Expected\n%v\nGot\n%v\n", expect[i].Data, fields[i].Data)
			}
		}
	}

	// Test load last
	fields := fieldDB.LoadLast(int(fieldDB.simID))
	expect := []Field{
		concentration, temperature,
	}
	for i := range fields {
		if !pfutil.CmplxEqualApprox(fields[i].Data, expect[i].Data, 1e-10) {
			t.Errorf("Expected\n%v\nGot\n%v\n", expect[i].Data, fields[i].Data)
		}
	}
}
