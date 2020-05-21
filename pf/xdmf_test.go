package pf

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestTopology(t *testing.T) {
	topo := XDMFTopology{Name: "topo", Type: "3DCoRectMesh", Dimensions: "100 500 500"}
	s := ""
	buf := bytes.NewBufferString(s)
	enc := xml.NewEncoder(buf)
	enc.Encode(topo)
	expect := "<Topology name=\"topo\" TopologyType=\"3DCoRectMesh\" Dimensions=\"100 500 500\"></Topology>"
	s = buf.String()
	if expect != s {
		t.Errorf("Expected\n%s\nGot\n%s\n", expect, s)
	}
}

func TestDataItem(t *testing.T) {
	data := XDMFDataItem{Dim: 2, Value: "0.0 0.0", Format: "XML"}
	s := ""
	buf := bytes.NewBufferString(s)
	enc := xml.NewEncoder(buf)
	enc.Encode(data)
	s = buf.String()
	expect := "<DataItem Dimension=\"2\" Format=\"XML\">0.0 0.0</DataItem>"
	if s != expect {
		t.Errorf("Expected\n%s\nGot\n%s\n", expect, s)
	}
}

func TestGeometry(t *testing.T) {
	geo := XDMFGeometry{
		Name:      "geo",
		DataItems: []XDMFDataItem{{}, {}},
	}
	s := ""
	buf := bytes.NewBufferString(s)
	enc := xml.NewEncoder(buf)
	enc.Encode(geo)
	s = buf.String()
	expect := "<Geometry name=\"geo\"><DataItem></DataItem><DataItem></DataItem></Geometry>"
	if s != expect {
		t.Errorf("Expected\n%s\nGot\n%s\n", s, expect)
	}
}

func TestCreateXDMF(t *testing.T) {
	fieldNames := []string{"conc", "eta"}
	xdmf := CreateXDMF(fieldNames, "myprefix", 2, []int{128, 128})
	s := ""
	buf := bytes.NewBufferString(s)
	enc := xml.NewEncoder(buf)
	enc.Encode(xdmf)
	s = buf.String()
}
