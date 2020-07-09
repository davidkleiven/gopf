package pf

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/cmplx"
	"os"

	"github.com/davidkleiven/gopf/pfutil"
)

// SDDTimeConstants is a struct for storing time-constants for the additional equations
// used in SDD
type SDDTimeConstants struct {
	// Orientation is the time constant used to evolve the orientation vector equation
	Orientation float64

	// DimerLength is the time constant used to evolve the length of the dimer
	// (e.g. l(t) = l_0 exp(-t/tau))
	DimerLength float64
}

// SDDMonitor is a type that is used to monitor the progress of the SDD stepper
type SDDMonitor struct {
	// MaxForce is the maximum fourier transformed force on the center
	MaxForce float64

	// ForcePowerSpectrum contains the power spectrum of the force
	// sum_{m=0}^{N-1} F(k_m)/N, where N is the number of nodes. This quantity
	// is the same as the real space integram
	// sum_ {m=0}^{N-1} F(k_m)/N = sum_{m=0}^{N-1} f(r_m), where F(k) = FFT(f(r))
	ForcePowerSpectrum float64

	// MaxTorque is the maximum torque exerted on the dimer
	MaxTorque float64

	// FieldNorm is the L2 norm of the fields which is given by
	// sum_{f=0}^{M_f-1} sum_{n=0}^{N-1} | Field[f].Data[n] |^2,
	// where M_f is the number of fields and N is the number of nodes
	FieldNorm float64

	// FieldNormChange contains the difference in L2 norm of between two
	// sucessive time steps
	FieldNormChange float64

	// LogFile is a writable file where the values of the quantities
	// above will be logged at each timestep. If not given (or nil),
	// no information will be logged
	LogFile *os.File
}

// Log prints the solver to a csv based format. This method can be attached
// as a callback to the solver
func (sddm *SDDMonitor) Log(s *Solver, epoch int) {
	if sddm.LogFile == nil {
		return
	}

	stat, err := sddm.LogFile.Stat()
	if err != nil {
		log.Fatalf("Error when retriieving stat: %s\n", err)
		return
	}
	size := stat.Size()
	header := []string{"MaxForce", "MaxTorque", "ForcePowerSpect", "FieldNorm", "FieldNormChange"}
	writer := csv.NewWriter(sddm.LogFile)
	defer writer.Flush()

	if size == 0 {
		writer.Write(header)
	}

	record := []string{
		fmt.Sprintf("%f", sddm.MaxForce),
		fmt.Sprintf("%f", sddm.MaxTorque),
		fmt.Sprintf("%f", sddm.ForcePowerSpectrum),
		fmt.Sprintf("%f", sddm.FieldNorm),
		fmt.Sprintf("%f", sddm.FieldNormChange),
	}
	writer.Write(record)
}

// SDD implements Shrinking-Dimer-Dynamics
type SDD struct {
	TimeConstants SDDTimeConstants

	// Alpha contains a weight for how to evaluate the force at the center of the dimer
	// the "end points" are given by
	// x_1 = x - alpha*l*v/2 and x_2 = x + alpha*l*v/2
	// where l is the length of the dimer and v is a unit orientation vector
	// The force exerted at the center is given by
	// F(x) = alpha*F(x_1) + (1-alpha)*F(x_2). Default when the SDD is constructed
	// from NewSDD is alpha = 0.5.
	Alpha float64

	// Dt is the timestep used to evlolve the equation
	Dt float64

	// CurrentStep contains the current starting iteration
	CurrentStep int

	// Minimum dimer length (default 0). When the length of dimer reaches this value
	// it will stop to shrink and maintain the minimum length
	MinDimerLength float64

	// Monitor holds data on the status of the SDD stepper
	Monitor SDDMonitor

	// InitDimerLength is the length of the dimer at the first timestep.
	// Default is to use the length of the orientation vector passed to
	// Init or SetOrientationVector
	InitDimerLength float64

	orientation []float64
	ft          *FFTWWrapper
	initialized bool
}

// NewSDD initializes a new SDD struct
func NewSDD(domainSize []int, model *Model) SDD {
	return SDD{
		TimeConstants: SDDTimeConstants{1.0, 1.0},
		Alpha:         0.5,
		orientation:   make([]float64, model.NumNodes()*len(model.Fields)),
		ft:            NewFFTW(domainSize),
		initialized:   false,
	}
}

func (sdd *SDD) checkTimeStep() {
	if sdd.Dt < 1e-16 {
		panic("Timestep not set in SDD. Make sure that the Dt attribute has explicitly been set.")
	}
}

// fft fourier transform all fields in the model. All derived fields are
// correctly updated prior to fourier transforming
func (sdd *SDD) fft(m *Model) {
	m.SyncDerivedFields()
	for _, f := range m.Fields {
		sdd.ft.FFT(f.Data)
	}
	for _, f := range m.DerivedFields {
		sdd.ft.FFT(f.Data)
	}
}

// ifft inverse fourier transforms the fields in the model
func (sdd *SDD) ifft(m *Model) {
	for _, f := range m.Fields {
		sdd.ft.IFFT(f.Data)
		pfutil.DivRealScalar(f.Data, float64(len(f.Data)))
	}
}

// Step evolves the system one step using semi-implicit euler
func (sdd *SDD) Step(m *Model) {
	if !sdd.initialized {
		panic("SDD: The method have to be initialized first. See SDD.Init\n")
	}
	sdd.checkTimeStep()

	fnorm := sdd.FieldNorm(m.Fields)
	sdd.Monitor.FieldNormChange = sdd.Monitor.FieldNorm - fnorm
	sdd.Monitor.FieldNorm = fnorm

	ftOrientation := make([]complex128, m.NumNodes()*len(m.Fields))
	for i := range sdd.orientation {
		ftOrientation[i] = complex(sdd.orientation[i], 0.0)
	}

	for i := range m.Fields {
		sdd.ft.FFT(ftOrientation[i*m.NumNodes() : (i+1)*m.NumNodes()])
	}

	rhsStart := make([]complex128, len(ftOrientation))
	rhsEnd := make([]complex128, len(ftOrientation))
	l := sdd.DimerLength(sdd.GetTime())

	// Shift to start
	sdd.ShiftFieldsAlongDimer(m.Fields, -0.5*l)
	sdd.fft(m)
	sdd.extractRHS(m, rhsStart)

	// Shift to end
	sdd.ifft(m)
	sdd.ShiftFieldsAlongDimer(m.Fields, l)
	sdd.fft(m)
	sdd.extractRHS(m, rhsEnd)

	// Shift back to center
	sdd.ifft(m)
	sdd.ShiftFieldsAlongDimer(m.Fields, -0.5*l)
	sdd.fft(m)

	work := make([]complex128, m.NumNodes())

	// Array that is used to track changes in the produced field
	origField := make([]complex128, m.NumNodes())
	sdd.Monitor.MaxForce = 0.0
	sdd.Monitor.ForcePowerSpectrum = 0.0
	cDt := complex(sdd.Dt, 0.0)

	for i := range m.Fields {
		copy(origField, m.Fields[i].Data)
		// Fill the work array with the weighted average of the field values
		w1 := complex(sdd.Alpha, 0.0)
		w2 := complex(1.0-sdd.Alpha, 0.0)
		for j := range work {
			work[j] = w1*rhsStart[j] + w2*rhsEnd[j]
		}
		d := m.Fields[i].Data
		activeFtOrient := ftOrientation[i*m.NumNodes() : (i+1)*m.NumNodes()]
		sdd.householder(work, activeFtOrient, 2.0, m.NumNodes())

		// Apply euler scheme
		for j := range d {
			d[j] = (d[j] + cDt*work[j])
		}
		work = m.GetDenum(i, sdd.ft.Freq, sdd.GetTime())
		dsm := sdd.householderDenum(work, activeFtOrient, 2.0)
		dsm.dot(d)

		for j := range d {
			diff := cmplx.Abs((d[j] - origField[j]) / cDt)
			sdd.Monitor.ForcePowerSpectrum += diff * diff / float64(m.NumNodes())
			if diff > sdd.Monitor.MaxForce {
				sdd.Monitor.MaxForce = diff
			}
		}
	}

	// Normalize power spectrum by <number of fields> * <number of nodes>
	sdd.Monitor.ForcePowerSpectrum /= float64(len(m.Fields) * m.NumNodes())
	sdd.Monitor.ForcePowerSpectrum = math.Sqrt(sdd.Monitor.ForcePowerSpectrum)

	// Update the right hand side such that it now contains rhsStart - rhsEnd
	torque := rhsStart // Torque on the dimer. Note to avoid re-allocation, torque shares the same underlying memory as rhsStart
	for i := range rhsStart {
		torque[i] = rhsStart[i] - rhsEnd[i]
	}

	// Add contribution from the linear part. The fourier transformed linear part is given
	// by F(k)*x, where x is the field. Thus the difference between the start and end is
	// F(k)*(x - alpha*l*v) - F(k)*(x + (1-alpha)*l*v) = -F(k)*v*l
	outerIdx := 0
	for i := range m.Fields {
		denum := m.GetDenum(i, sdd.ft.Freq, sdd.GetTime())
		for j := range denum {
			torque[outerIdx] -= denum[j] * ftOrientation[outerIdx] * complex(l, 0.0)
			outerIdx++
		}
	}

	// UpdateOrientation updates the orientation. The dynamic equation for the orientation is
	//
	//   dv        (I - vv^T)*(F_1 - F_2)
	// ------ =  --------------------------
	//   dt               l(t)*tau
	//
	// where F_1 and F_2 are forces at the current end images. The fourier transformed
	// F_1 - F_2 should be passed in ftForce. v is the orientation vector, I is the identify matrix.
	// l(t) is the length of dimer at time t and tau is a time constant (see SDD.TimeConstants.Orientation)
	sdd.householder(torque, ftOrientation, 1.0, m.NumNodes())

	// Inverse FFT the rhs
	for i := range m.Fields {
		sdd.ft.IFFT(torque[i*m.NumNodes() : (i+1)*m.NumNodes()])
	}
	pfutil.DivRealScalar(torque, float64(m.NumNodes()))

	// Sanity check: orientation should be orthogonal to rhs
	rhsDotOrientation := 0.0
	for i := range torque {
		rhsDotOrientation += real(torque[i]) * sdd.orientation[i]
	}

	if math.Abs(rhsDotOrientation) > 1e-10 {
		log.Printf("Warning! Torque is not orthogonal to the current orientation. (%.2e)\n", rhsDotOrientation)
	}

	// Update the orientation
	sdd.Monitor.MaxTorque = 0.0
	for i := range sdd.orientation {
		sdd.orientation[i] -= sdd.Dt * real(torque[i]) / (sdd.DimerLength(sdd.GetTime()) * sdd.TimeConstants.Orientation)
		if math.Abs(real(torque[i])) > sdd.Monitor.MaxTorque {
			sdd.Monitor.MaxTorque = math.Abs(real(torque[i]))
		}
	}

	// Normalize the orientation vector
	length := math.Sqrt(pfutil.Dot(sdd.orientation, sdd.orientation))
	for i := range sdd.orientation {
		sdd.orientation[i] /= length
	}

	sdd.ifft(m)
	sdd.CurrentStep++
}

func (sdd *SDD) extractRHS(m *Model, rhs []complex128) {
	for i := range m.Fields {
		newRHS := m.GetRHS(i, sdd.ft.Freq, sdd.GetTime())
		copy(rhs[i*m.NumNodes():], newRHS)
	}
}

// householder applies the Householder reflection to the data array. data is the fourier transformed
// vector where the operation should be applied and ftOrientation is the fourier traansformed
// orientation vector
func (sdd *SDD) householder(data []complex128, ftOrientation []complex128, sigma float64, numNodes int) {
	dataDotOrientation := complex(0.0, 0.0)
	for i := range data {
		dataDotOrientation += data[i] * cmplx.Conj(ftOrientation[i])
	}
	dataDotOrientation /= complex(float64(numNodes), 0.0)

	cSigma := complex(sigma, 0.0)
	for i := range data {
		data[i] = data[i] - cSigma*ftOrientation[i]*dataDotOrientation
	}
}

// householderDenum returns the result of (I - dt*(I-vv^T))^-1. Internally, the Sherman-Morrison formula is used
func (sdd *SDD) householderDenum(data []complex128, ftOrientation []complex128, sigma float64) diagonalShermannMorrison {
	cDt := complex(sdd.Dt, 0.0)

	dsm := diagonalShermannMorrison{
		invDiagonal: make([]complex128, len(data)),
		u:           make([]complex128, len(data)),
		v:           make([]complex128, len(data)),
	}

	cSigma := complex(sigma, 0.0)
	for i := range data {
		dsm.invDiagonal[i] = 1.0 / (1.0 - cDt*data[i])
		dsm.v[i] = cSigma * data[i] * cDt * cmplx.Conj(ftOrientation[i]) / complex(float64(len(data)), 0.0)
	}
	dsm.u = ftOrientation
	return dsm
}

// ShiftFieldsAlongDimer shifts the field along the dimer. The length is given by scale*v,
// where v is the dimer orientation vector
func (sdd *SDD) ShiftFieldsAlongDimer(fields []Field, scale float64) {
	outerIdx := 0
	for i := range fields {
		for j := range fields[i].Data {
			fields[i].Data[j] += complex(scale*sdd.orientation[outerIdx], 0.0)
			outerIdx++
		}
	}
}

// GetTime returns the current time
func (sdd *SDD) GetTime() float64 {
	return float64(sdd.CurrentStep) * sdd.Dt
}

// DimerLength returns the the length of the timer at the given time
func (sdd *SDD) DimerLength(t float64) float64 {
	l := sdd.InitDimerLength * math.Exp(-t/sdd.TimeConstants.DimerLength)
	if l < sdd.MinDimerLength {
		return sdd.MinDimerLength
	}
	return l
}

// RequiredDimerLengthTime returns the the time needed to reach the passed length
func (sdd *SDD) RequiredDimerLengthTime(l float64) float64 {
	return sdd.TimeConstants.DimerLength * math.Log(sdd.InitDimerLength/l)
}

// FieldNorm calculates the L2 norm of the fields at the current saddle
func (sdd *SDD) FieldNorm(fields []Field) float64 {
	norm := 0.0
	for _, f := range fields {
		for j := range f.Data {
			v := cmplx.Abs(f.Data[j])
			norm += v * v
		}
	}
	return norm
}

// Init initializes the orientation vector and the dimer length
func (sdd *SDD) Init(init []Field, final []Field) {
	outerIdx := 0
	sdd.InitDimerLength = 0.0
	for i := range init {
		for j := range init[i].Data {
			sdd.orientation[outerIdx] = real(final[i].Data[j] - init[i].Data[j])
			sdd.InitDimerLength += sdd.orientation[outerIdx] * sdd.orientation[outerIdx]
			outerIdx++
		}
	}

	sdd.InitDimerLength = math.Sqrt(sdd.InitDimerLength)

	// Normalize the vector
	for i := range sdd.orientation {
		sdd.orientation[i] /= sdd.InitDimerLength
	}
	sdd.initialized = true
}

// SetInitialOrientation sets the starting orientation vector
// the initial dimer length is set to the length of the passed orientation vector
// Subsequently, the passed orientation vector will be normalized to unit length
func (sdd *SDD) SetInitialOrientation(orient []float64) {
	if len(orient) != len(sdd.orientation) {
		panic("Inconsistent length of the passed orientaiton vector")
	}
	sdd.InitDimerLength = 0.0
	for i := range orient {
		sdd.InitDimerLength += orient[i] * orient[i]
	}
	sdd.InitDimerLength = math.Sqrt(sdd.InitDimerLength)
	copy(sdd.orientation, orient)
	for i := range sdd.orientation {
		sdd.orientation[i] /= sdd.InitDimerLength
	}
	sdd.initialized = true
}

// SetFilter is implemented to satisfy the TimeStepper interface. However, a call to this method
// will panic because model filters are currently not supported by this stepper.
func (sdd *SDD) SetFilter(filter ModalFilter) {
	panic("SDD: Does not support modal filters")
}

// SaveOrientation stores the current orientation vector to a csv file
func (sdd *SDD) SaveOrientation(fname string) {
	numNodes := pfutil.ProdInt(sdd.ft.Dimensions)
	numFields := len(sdd.orientation) / numNodes
	data := make([]CsvData, numFields)
	for i := range data {
		data[i] = CsvData{
			Name: fmt.Sprintf("orientation%d", i),
			Data: &pfutil.RealSlice{Data: sdd.orientation[i*numNodes : (i+1)*numNodes]},
		}
	}

	SaveCsv(fname, data, sdd.ft.Dimensions)
}

// diagonalShermannMorrison is a type that stores information to carry out inner products of the form
// (D + uv^T)^-1 * b, where b is an arbitrary vector. According to the Shermann-Morrison formula, the
// inverse matrix is given by D^{-1} - D^{-1}uv^TD^{-1}/(1 + u^TD^{-1}v)
type diagonalShermannMorrison struct {
	invDiagonal []complex128
	u           []complex128
	v           []complex128
}

// dot performed the dot product between the matrix represented by the shermann morrison
func (dsm *diagonalShermannMorrison) dot(vec []complex128) {
	denum := complex(1.0, 0.0)
	vDotInvDDotVec := complex(0.0, 0.0)
	for i := range dsm.invDiagonal {
		denum += dsm.u[i] * dsm.invDiagonal[i] * dsm.v[i]
		vDotInvDDotVec += dsm.v[i] * dsm.invDiagonal[i] * vec[i]
	}

	for i := range dsm.invDiagonal {
		vec[i] = dsm.invDiagonal[i]*vec[i] - dsm.invDiagonal[i]*dsm.u[i]*vDotInvDDotVec/denum
	}
}
