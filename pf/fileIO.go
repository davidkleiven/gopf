package pf

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/davidkleiven/gopf/pfutil"
)

// Uint8IO is a struct used to store fields as uint8
type Uint8IO struct {
	Prefix string
}

// NewUint8IO returns a new Uint8IO instance
func NewUint8IO(prefix string) Uint8IO {
	return Uint8IO{
		Prefix: prefix,
	}
}

// SaveFields can be passed as a callback to the solver. It stores each
// field in a raw binary file.
func (u *Uint8IO) SaveFields(s *Solver, epoch int) {
	for _, f := range s.Model.Fields {
		fname := fmt.Sprintf("%s_%s_%d.bin", u.Prefix, f.Name, epoch)
		min := pfutil.MinReal(f.Data)
		max := pfutil.MaxReal(f.Data)
		uint8Rep := RealPartAsUint8(f.Data, min, max)
		out, err := os.Create(fname)
		if err != nil {
			panic(err)
		}
		binary.Write(out, binary.BigEndian, uint8Rep)
		out.Close()
	}
}

// Float64IO stores the fields as raw binary files using BigEndian. The datatype is
// float64
type Float64IO struct {
	Prefix string
}

// NewFloat64IO returns a new Float64IO. All files are prepended ay prefix
func NewFloat64IO(prefix string) Float64IO {
	return Float64IO{Prefix: prefix}
}

// SaveFields stores all fields as raw binary files. It can be passed as a callback to the
// solver
func (fl *Float64IO) SaveFields(s *Solver, epoch int) {
	for _, f := range s.Model.Fields {
		fname := fmt.Sprintf("%s_%s_%d.bin", fl.Prefix, f.Name, epoch)
		f.SaveReal(fname)
	}
}

// LoadFloat64 loads an array of float64 encoded as binary data
// it is assumed that the it is stored with BigEndian
func LoadFloat64(fname string) []float64 {
	infile, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	stats, err := infile.Stat()
	if err != nil {
		panic(err)
	}
	size := stats.Size()
	bytes := make([]byte, size)

	data := make([]float64, len(bytes)/8)
	binary.Read(infile, binary.BigEndian, data)
	return data
}

// SaveFloat64 writes a float sice to a binary file. BigEndian is used.
// Files stored with this function can be read using LoadFloat64
func SaveFloat64(fname string, data []float64) {
	out, err := os.Create(fname)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	defer out.Close()
	binary.Write(out, binary.BigEndian, data)
}
