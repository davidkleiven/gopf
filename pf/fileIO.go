package pf

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

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

// CsvIO writes data to text csv text files
type CsvIO struct {
	Prefix     string
	DomainSize []int
}

// SaveFields stores the results in CSV files. The format
// X, Y, Z, field1, field2, field3
// etc.
func (cio *CsvIO) SaveFields(s *Solver, epoch int) {
	if cio.DomainSize == nil {
		panic("Domain size not given. Data will not be written to file\n")
	}

	if pfutil.ProdInt(cio.DomainSize) != s.Model.NumNodes() {
		msg := fmt.Sprintf("Inconsistent domain size. Expected %d nodes, got %d\n", s.Model.NumNodes(), pfutil.ProdInt(cio.DomainSize))
		panic(msg)
	}

	header := []string{"X", "Y", "Z"}
	for _, f := range s.Model.Fields {
		header = append(header, f.Name)
	}

	fname := cio.Prefix + fmt.Sprintf("_%d.csv", epoch)
	out, err := os.Create(fname)
	if err != nil {
		log.Fatalf("Could not open file: %s\n", err)
		return
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	defer writer.Flush()

	writer.Write(header)
	record := make([]string, len(header))
	pos := make([]int, 3)
	for i := 0; i < s.Model.NumNodes(); i++ {
		position := pfutil.Pos(cio.DomainSize, i)
		copy(pos, position)
		for j := 0; j < 3; j++ {
			record[j] = fmt.Sprintf("%d", pos[j])
		}
		for j, f := range s.Model.Fields {
			record[j+3] = fmt.Sprintf("%f", real(f.Data[i]))
		}
		writer.Write(record)
	}
}

// LoadCSV loads data from CSV file and returns an array of fields
func LoadCSV(fname string) []Field {
	infile, err := os.Open(fname)
	if err != nil {
		msg := fmt.Sprintf("Could not open file: %s\n", err)
		panic(msg)
	}
	defer infile.Close()

	reader := csv.NewReader(infile)
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Could not read header: %s\n", header)
		return nil
	}

	fields := make([]Field, len(header)-3)
	for i := range fields {
		fields[i].Name = header[i+3]
		fields[i].Data = []complex128{}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			return fields
		} else if err != nil {
			log.Fatalf("Could not read record: %s\n", err)
		}

		for i := 0; i < len(fields); i++ {
			v, _ := strconv.ParseFloat(record[i+3], 64)
			fields[i].Data = append(fields[i].Data, complex(v, 0.0))
		}
	}
}
