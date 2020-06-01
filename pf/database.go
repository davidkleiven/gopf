package pf

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davidkleiven/gopf/pfutil"
)

// FieldDB is a type used for storing field data in a SQl database. A field database
// consists of the following tables
//
// 1. positions (id int, x int, y int, z int)
// 		Describes the positions in 3D space of all nodes. The first column is the node
// 		number in the simulation cell and the remaining columns represent the x, y and
// 		z index, respectively. If the simulation domain is 2D, the z column is always
//		zero.
//
// 2. fields (id int, name text, value real, positionId int, timestep int, simID int)
//		Describes the value of the fields at a given timestep and point
// 		- id: Row identifier that is auto-incremented when new rows are added
//		- name: Name of the field
//		- value: Value of the field
//		- positionId: ID pointing to the positions table. The position in 3D space
// 			of the node represented by this row, is given by the corresponding row
// 			in the positions table
//		- timestep: Timestep of the record
//		- simID: Unique ID identifying all entries in the database written by the
//			current simulation
//
// 3. simAttributes (key text, value float, simID int)
//		Additional attributes that belongs to the simulations. Input arguments can
// 		for example be stored here.
//		- key: Name of the attribute
//		- value: Value of the attribute
// 		- simID: Unique ID which is common to all entries written by the current
//			simulation
//
// 4. comments (simID int, value text)
//		Describes comments about the simulation.
//		- simID: Unique ID which is common to all entries written by the current
//			simulation
//		- value: A text string describing the simulation
//
// 5. simIDs (simID int, creationTime text)
//		List of all simulation IDs in the database
//		- simID: Unique ID
//		- creationTime: Timestamp for when the simulation ID was created
//
// 6. simTextAttributes (key TEXT, value TEXT, simID int)
//		Same as simAttributes, apart from the value field is a string.
//
// 7. timeseries (key TEXT, value float, timestep int, simID int)
//		Describes time varying data typically derived from the fields in the
//		calculations. Some examples: peak concentration in a diffusion calculation,
//		peak stress in an elasiticy calculation, average domain size in a spinoidal
//		calculation etc.
//		- key: Name of the item
//		- value: Value of the data point
//		- timestep: Timestep
//		- simID: Unique ID which is common to all entries written by the current
//			simulation
type FieldDB struct {
	DB *sql.DB

	// DomainSize of the simulation domain. This is needed in order to populate the
	// position table
	DomainSize []int

	// simId is a random integer that is used to identify all items inserted in the
	// current run
	simID int32

	// true if the database has been initialized
	initialized bool
}

// initialize runs a set of sql statements to build the database. Note that subsequent calls
// to this function has no effect
func (fdb *FieldDB) initialize() {
	statement, err := fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS positions (id INTEGER PRIMARY KEY, X INTEGER, Y INTEGER, Z INTEGER)")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS fields (id INTEGER PRIMARY KEY, name TEXT, value REAL, positionId INTEGER, timestep INTEGER, simID INTEGER, FOREIGN KEY(positionId) REFERENCES positions(id))")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS simAttributes (key TEXT, value REAL, simID INTEGER, FOREIGN KEY(simID) REFERENCES fields(simID))")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS comments (simID INTEGER, value TEXT)")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS simTextAttributes (key TEXT, value TEXT, simID INTEGER, FOREIGN KEY(simID) REFERENCES fields(simID))")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS timeseries (key TEXT, value REAL, timestep INTEGER, simID INTEGER)")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	statement, err = fdb.DB.Prepare("CREATE TABLE IF NOT EXISTS simIDs (simID INTEGER UNIQUE, creationTime TEXT)")
	if err != nil {
		panic(err)
	}
	statement.Exec()

	if fdb.simID == 0 {
		source := rand.NewSource(time.Now().UnixNano())
		generator := rand.New(source)
		fdb.simID = generator.Int31()

		// Insert into the database
		statement, err = fdb.DB.Prepare("INSERT INTO simIDs (simID, creationTime) VALUES (?, datetime('now', 'localtime'))")
		if err != nil {
			panic(err)
		}
		statement.Exec(fdb.simID)
	}

	fdb.initialized = true
}

// positionTableIsPopulated return true if the position table has been populated
func (fdb *FieldDB) positionTableIsPopulated() bool {
	result, err := fdb.DB.Query("SELECT COUNT(*) FROM positions")
	if err != nil {
		panic(err)
	}
	var numRows int
	for result.Next() {
		result.Scan(&numRows)
	}
	return numRows > 1
}

// populatePositionTables inserts values into the position table
func (fdb *FieldDB) populatePositionsTable() {
	pos3 := make([]int, 3)
	numNodes := pfutil.ProdInt(fdb.DomainSize)
	tx, err := fdb.DB.Begin()

	if err != nil {
		panic(err)
	}
	for i := 0; i < numNodes; i++ {
		pos := pfutil.Pos(fdb.DomainSize, i)
		copy(pos3, pos)
		statement, err := tx.Prepare("INSERT INTO positions (X, Y, Z) VALUES (?, ?, ?)")

		if err != nil {
			tx.Rollback()
			return
		}
		_, err = statement.Exec(pos3[0], pos3[1], pos3[2])

		if err != nil {
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// insertRealPart inserts the real part of a set of field values into the database
func (fdb *FieldDB) insertRealPart(name string, timestep int, values []complex128) {
	if !fdb.domainSizeOk() {
		panic("FieldDB: Domain size does not match the positions table in the database")
	}

	if len(values) != pfutil.ProdInt(fdb.DomainSize) {
		panic("FieldDB: The passed array does not match the specified domain size")
	}
	tx, _ := fdb.DB.Begin()
	for i := range values {
		statement, err := tx.Prepare("INSERT INTO fields (name, value, positionId, timestep, simID) VALUES (?, ?, ?, ?, ?)")

		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}

		_, err = statement.Exec(name, real(values[i]), i, timestep, fdb.simID)

		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// SaveFields stores all the field to the database. This function satisfies the
// SolverCB type, and can thus be attached as a callback to a solver
func (fdb *FieldDB) SaveFields(s *Solver, epoch int) {
	if !fdb.initialized {
		fdb.initialize()
	}

	if !fdb.positionTableIsPopulated() {
		fdb.populatePositionsTable()
	}

	for _, f := range s.Model.Fields {
		fdb.insertRealPart(f.Name, epoch, f.Data)
	}
}

// Comment adds a comment associated with the current simulation ID
func (fdb *FieldDB) Comment(comment string) {
	if !fdb.initialized {
		fdb.initialize()
	}
	statement, err := fdb.DB.Prepare("INSERT INTO comments (simID, value) VALUES (?, ?)")
	if err != nil {
		log.Fatal(err)
		return
	}
	statement.Exec(fdb.simID, comment)
}

// SetAttr adds a set of key-value pairs associated with the current simID
func (fdb *FieldDB) SetAttr(attr map[string]float64) {
	if !fdb.initialized {
		fdb.initialize()
	}

	tx, err := fdb.DB.Begin()
	if err != nil {
		log.Fatal(err)
		return
	}

	for k, v := range attr {
		statement, err := tx.Prepare("INSERT INTO simAttributes (key, value, simID) VALUES (?, ?, ?)")
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
		_, err = statement.Exec(k, v, fdb.simID)
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// SetTextAttr sets text attributes associated with the current simulation
func (fdb *FieldDB) SetTextAttr(attr map[string]string) {
	if !fdb.initialized {
		fdb.initialize()
	}
	tx, err := fdb.DB.Begin()
	if err != nil {
		log.Fatal(err)
		return
	}

	for k, v := range attr {
		statement, err := tx.Prepare("INSERT INTO simTextAttributes (key, value, simID) VALUES (?, ?, ?)")
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}

		_, err = statement.Exec(k, v, fdb.simID)
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// TimeSeries inserts data into the timeseries table
func (fdb *FieldDB) TimeSeries(data map[string]float64, timestep int) {
	if !fdb.initialized {
		fdb.initialize()
	}
	tx, err := fdb.DB.Begin()
	if err != nil {
		log.Fatal(err)
		return
	}

	for k, v := range data {
		statement, err := tx.Prepare("INSERT INTO timeseries (key, value, timestep, simID) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
		_, err = statement.Exec(k, v, timestep, fdb.simID)
		if err != nil {
			log.Fatal(err)
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// domainSizeOk returns true if the domain size stored locally matches the one
// in the database. If domainSize is nil, this function always return true.
func (fdb *FieldDB) domainSizeOk() bool {
	if fdb.DomainSize == nil {
		return true
	}

	columns := []string{"X", "Y", "Z"}

	for i := 0; i < len(fdb.DomainSize); i++ {
		sqlStr := fmt.Sprintf("SELECT MAX(%s) FROM positions", columns[i])
		rows, err := fdb.DB.Query(sqlStr)
		if err != nil {
			log.Fatal(err)
			return false
		}

		var maxval int
		for rows.Next() {
			rows.Scan(&maxval)
		}
		if maxval != fdb.DomainSize[i]-1 {
			return false
		}
	}
	return true
}
