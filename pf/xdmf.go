package pf

import (
	"encoding/xml"
	"fmt"
	"os"
)

// XDMFTopology is a type used to represent the Topology item in paraview xdmf format
type XDMFTopology struct {
	XMLName    xml.Name `xml:"Topology"`
	Name       string   `xml:"name,attr,omitempty"`
	Type       string   `xml:"TopologyType,attr,omitempty"`
	Dimensions string   `xml:"Dimensions,attr,omitempty"`
	Reference  string   `xml:"Reference,attr,omitempty"`
}

// XDMFGeometry is a structure used to represent eh Geometry item in paraview xdmf format
type XDMFGeometry struct {
	XMLName   xml.Name       `xml:"Geometry"`
	Name      string         `xml:"name,attr,omitempty"`
	DataItems []XDMFDataItem `xml:"DataItem"`
	Type      string         `xml:"Type,attr,omitempty"`
	Reference string         `xml:"Reference,attr,omitempty"`
}

// XDMFDataItem is a type used to represent the data item in paraview xdmf format
type XDMFDataItem struct {
	XMLName    xml.Name `xml:"DataItem"`
	Dim        int      `xml:"Dimension,attr,omitempty"`
	Format     string   `xml:"Format,attr,omitempty"`
	Value      string   `xml:",chardata"`
	NumberType string   `xml:"NumberType,attr,omitempty"`
	Dimensions string   `xml:"Dimensions,attr,omitempty"`
	Precision  int      `xml:"Precision,attr,omitempty"`
	DataType   string   `xml:"DataType,attr,omitempty"`
	Endian     string   `xml:"Endian,attr,omitempty"`
}

// XDMFAttribute is type used to represent the attribute item in paraview xdmf format
type XDMFAttribute struct {
	XMLName  xml.Name `xml:"Attribute"`
	DataItem XDMFDataItem
	Name     string `xml:"Name,attr,omitempty"`
	Center   string `xml:"Center,attr,omitempty"`
}

// XDMFTime is used to represent the time attribute in paraview xfdmf format
type XDMFTime struct {
	XMLName  xml.Name `xml:"Time"`
	Type     string   `xml:"TimeType,attr"`
	DataItem XDMFDataItem
}

// XDMFGrid is used to represent the grid structure in paraview xdmf format
type XDMFGrid struct {
	XMLName        xml.Name        `xml:"Grid"`
	Name           string          `xml:"Name,attr,omitempty"`
	Type           string          `xml:"GridType,attr,omitempty"`
	CollectionType string          `xml:"CollectionType,attr,omitempty"`
	Grids          []XDMFGrid      `xml:"Grid"`
	Attributes     []XDMFAttribute `xml:"Attribute"`
	Topology       XDMFTopology    `xml:"Topology,omitempty"`
	Geometry       XDMFGeometry    `xml:"Geometry,omitempty"`
	Time           *XDMFTime       `xml:"Time"`
}

// XDMFDomain represent domain attribute in paraview xdmf format
type XDMFDomain struct {
	XMLName  xml.Name `xml:"Domain"`
	Topology XDMFTopology
	Geometry XDMFGeometry
	Grid     XDMFGrid
}

// XDMF represents the domain attribute in paraview xdmf format
type XDMF struct {
	XMLName xml.Name `xml:"Xdmf"`
	Domain  XDMFDomain
}

// CreateXDMF returns a new instance of XDMF
func CreateXDMF(fieldNames []string, prefix string, num int, domainSize []int) XDMF {
	dim := len(domainSize)
	dimensions := ""
	geoType := ""
	if len(domainSize) == 2 {
		dimensions = fmt.Sprintf("%d %d", domainSize[0], domainSize[1])
		geoType = "ORIGIN_DXDY"
	} else if len(domainSize) == 3 {
		dimensions = fmt.Sprintf("%d %d %d", domainSize[0], domainSize[1], domainSize[2])
		geoType = "ORIGIN_DXDYDZ"
	} else {
		panic("Length of domain size to be either 2 or 3")
	}
	xdmf := XDMF{}
	xdmf.Domain.Topology = XDMFTopology{
		Name:       "topo",
		Type:       fmt.Sprintf("%dDCoRectMesh", dim),
		Dimensions: dimensions,
	}

	xdmf.Domain.Geometry = XDMFGeometry{
		Name: "geo",
		Type: geoType,
		DataItems: []XDMFDataItem{XDMFDataItem{
			Format:     "XML",
			Dimensions: fmt.Sprintf("%d", dim),
			Value:      "0.0 0.0 0.0",
		},
			XDMFDataItem{
				Format:     "XML",
				Dimensions: fmt.Sprintf("%d", dim),
				Value:      "1.0 1.0 1.0",
			},
		},
	}

	timeStr := ""
	for i := 0; i < num; i++ {
		timeStr += fmt.Sprintf("%f ", float64(i)/float64(num))
	}
	timeStr += fmt.Sprintf("%d", num)

	xdmf.Domain.Grid = XDMFGrid{
		Name:           "TimeSeries",
		Type:           "Collection",
		CollectionType: "Temporal",
		Grids:          make([]XDMFGrid, num),
		Time: &XDMFTime{Type: "HyperSlab",
			DataItem: XDMFDataItem{Format: "XML", NumberType: "Float", Dimensions: fmt.Sprintf("%d", dim), Value: timeStr},
		},
	}

	for i := 0; i < num; i++ {
		grid := XDMFGrid{
			Name:       fmt.Sprintf("T%d", i+1),
			Type:       "Uniform",
			Attributes: make([]XDMFAttribute, len(fieldNames)),
			Topology:   XDMFTopology{Reference: "/Xdmf/Domain/Topology[1]"},
			Geometry:   XDMFGeometry{Reference: "/Xdmf/Domain/Geometry[1]"},
		}
		for j := 0; j < len(fieldNames); j++ {
			grid.Attributes[j] = XDMFAttribute{
				Name:   fieldNames[j],
				Center: "Node",
				DataItem: XDMFDataItem{
					Format:     "Binary",
					DataType:   "Float",
					Endian:     "Big",
					Precision:  8,
					Dimensions: dimensions,
					Value:      fmt.Sprintf("%s_%s_%d.bin", prefix, fieldNames[j], i),
				},
			}
		}
		xdmf.Domain.Grid.Grids[i] = grid
	}
	return xdmf
}

// WriteXDMF creates a xdmf file that can be used by paraview
func WriteXDMF(fname string, fields []string, prefix string, num int, domainSize []int) {
	writer, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	enc := xml.NewEncoder(writer)
	enc.Indent("", "    ")
	enc.Encode(CreateXDMF(fields, prefix, num, domainSize))
	writer.Close()
}
