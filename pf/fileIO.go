package pf

import (
	"encoding/binary"
	"fmt"
	"os"
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
		min := MinReal(f.Data)
		max := MaxReal(f.Data)
		uint8Rep := RealPartAsUint8(f.Data, min, max)
		out, err := os.Create(fname)
		if err != nil {
			panic(err)
		}
		binary.Write(out, binary.LittleEndian, uint8Rep)
		out.Close()
	}
}
