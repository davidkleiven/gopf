package pf

import (
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"testing"

	"github.com/davidkleiven/gopf/pfutil"
	"gonum.org/v1/gonum/mat"
)

func TestFourierOuterProduct(t *testing.T) {
	// Write simple test case to verify that we can perform outer products in
	// the fourier domain
	N := 8
	data := make([]complex128, N*N)
	vector := make([]complex128, N*N)
	result := make([]float64, N*N)
	dataDot := 0.0
	for i := range data {
		data[i] = complex(float64(i)/10.0, 0.0)
		vector[i] = complex(float64(i*i-i)/10.0, 0.0)
		dataDot += real(data[i] * vector[i])
	}

	for i := range result {
		result[i] = real(vector[i]) * dataDot
	}

	// Perform the calculation using FFTs
	ft := NewFFTW([]int{N, N})
	ft.FFT(data)
	ft.FFT(vector)

	dotProd := complex(0.0, 0.0)
	for i := range data {
		dotProd += cmplx.Conj(vector[i]) * data[i]
	}
	dotProd /= complex(float64(len(data)), 0.0)

	for i := range data {
		vector[i] *= dotProd
	}
	ft.IFFT(vector)
	pfutil.DivRealScalar(vector, float64(len(vector)))

	tol := 1e-6
	for i := range vector {
		re := real(vector[i])
		im := imag(vector[i])
		if math.Abs(re-result[i]) > tol || math.Abs(im) > tol {
			t.Errorf("Expected (%f, 0.0) got (%f, %f)\n", result[i], re, im)
		}
	}
}

func TestDimerLengthTime(t *testing.T) {
	dimer := SDD{InitDimerLength: 5.0}
	reqTime := dimer.RequiredDimerLengthTime(0.5)
	length := dimer.DimerLength(reqTime)
	if math.Abs(length-0.5) > 1e-10 {
		t.Errorf("Expected 0.5 got %f\n", length)
	}
}

func TestDoubleWell(t *testing.T) {
	N := 4
	field := NewField("conc", N*N, nil)
	init := NewField("concInit", N*N, nil)
	final := NewField("concFinal", N*N, nil)
	expectSaddle := make([]complex128, N*N) // Should be a saddle at 0.0
	for i := range init.Data {
		init.Data[i] = complex(-1.0, 0.0)
		final.Data[i] = complex(1.5, 0.0)
		field.Data[i] = 0.5 * (init.Data[i] + final.Data[i])
	}

	model := NewModel()
	model.AddField(field)
	model.AddEquation("dconc/dt = conc - conc^3")

	dt := 0.1
	sdd := NewSDD([]int{N, N}, &model)
	sdd.Init([]Field{init}, []Field{final})
	sdd.InitDimerLength = 0.1
	solver := NewSolver(&model, []int{N, N}, dt)
	sdd.Dt = dt
	solver.Stepper = &sdd

	length := sdd.DimerLength(0.0)
	finalLength := 0.000001 * length
	finalTime := sdd.RequiredDimerLengthTime(finalLength)
	nsteps := int(finalTime/dt) + 1
	solver.Solve(1, nsteps)

	tol := 1e-6
	for i := range model.Fields[0].Data {
		if cmplx.IsNaN(model.Fields[0].Data[i]) {
			t.Errorf("NaN detected in solution\n%v\n", model.Fields[0].Data)
			return
		}
	}
	if !pfutil.CmplxEqualApprox(expectSaddle, model.Fields[0].Data, tol) {
		t.Errorf("Expected:\n%v\nGot\n%v\n", expectSaddle, model.Fields[0].Data)
	}
}

func Test2DSurface(t *testing.T) {
	// Locate the saddle points of the surface
	// E(x, y) = (x^2 - 1)^2 + y^2. There is a saddle point
	// at (0, 0)
	N := 1
	x := NewField("xCrd", N, nil)
	y := NewField("yCrd", N, nil)
	model := NewModel()
	model.AddField(x)
	model.AddField(y)
	model.AddScalar(NewScalar("FOUR", complex(4.0, 0.0)))
	model.AddScalar(NewScalar("TWO", complex(2.0, 0.0)))
	model.AddEquation("dxCrd/dt = FOUR*xCrd - FOUR*xCrd^3")
	model.AddEquation("dyCrd/dt = -TWO*yCrd")

	dt := 0.001
	solver := NewSolver(&model, []int{N, N}, dt)
	model.Summarize()
	stepper := NewSDD([]int{N, N}, &model)
	stepper.Dt = dt
	stepper.SetInitialOrientation([]float64{0.1, 0.5})
	x.Data[0] = complex(0.2, 0.0)
	y.Data[0] = complex(0.3, 0.0)
	solver.Stepper = &stepper

	finalLength := 1e-5
	maxSteps := int(stepper.RequiredDimerLengthTime(finalLength) / dt)
	solver.Solve(1, maxSteps)

	// There is a saddle point at (0, 0)
	xSaddle := real(x.Data[0])
	ySaddle := real(y.Data[0])
	fmt.Printf("%v\n", stepper.orientation)
	tol := 1e-3
	if math.Abs(xSaddle) > tol || math.Abs(ySaddle) > tol {
		t.Errorf("Expected (0, 0) got (%f, %f)\n", xSaddle, ySaddle)
	}

	if math.Abs(stepper.orientation[0]-1.0) > tol || math.Abs(stepper.orientation[1]) > tol {
		t.Errorf("Expected orientation (1, 0) got (%f, %f)\n", stepper.orientation[0], stepper.orientation[1])
	}
}

func insertCircleAtCenter(data []complex128, domainSize []int, radius int) {
	for i := range data {
		pos := pfutil.Pos(domainSize, i)
		dx := pos[0] - domainSize[0]/2
		dy := pos[1] - domainSize[1]/2
		rSq := dx*dx + dy*dy
		if rSq <= radius*radius {
			data[i] = complex(1.0, 0.0)
		}
	}
}

func TestClassicalNucleation(t *testing.T) {
	// Set to true in order to output files (useful for local debugging)
	storeFiles := false

	// Folder where the files will be stored in case storeFiles = true
	folder := "./"
	//folder := "/home/gudrun/davidkl/Documents/Dump/FileTesting/"
	N := 64
	field := NewField("phi", N*N, nil)
	init := NewField("phiInit", N*N, nil)
	final := NewField("phiFinal", N*N, nil)

	// Initialize fields
	for i := range init.Data {
		init.Data[i] = complex(-1.0, 0.0)
		final.Data[i] = complex(-1.0, 0.0)
		field.Data[i] = complex(-1.0, 0.0)
	}

	gamma := 0.5 // Gradient coefficient
	rho := 0.05  // Coefficient controlling the bulk driving force

	// Calculate the surface tension in the sharp interface limit
	// Cahn-Hilliard
	//                        * phi_1
	//                       *
	// sigma = sqrt(gamma/2) * dphi sqrt(f_s(phi))
	//                       *
	//                      * phi_0
	//
	// f(phi) = -phi^2/2 + phi^4/4 -rho*(3*phi - phi^3)/4
	//
	// Thus, the surface formation energy is given by

	integral := 2.0 / 3.0
	surfaceTension := math.Sqrt(gamma/2.0) * integral
	rc := 2.0 * surfaceTension / rho

	insertCircleAtCenter(final.Data, []int{N, N}, 15)
	insertCircleAtCenter(init.Data, []int{N, N}, 10)
	pfutil.Blur(&pfutil.RealPartSlice{Data: final.Data}, []int{N, N}, &pfutil.BoxKernel{Width: 2})
	pfutil.Blur(&pfutil.RealPartSlice{Data: init.Data}, []int{N, N}, &pfutil.BoxKernel{Width: 2})

	insertCircleAtCenter(field.Data, []int{N, N}, 12)
	pfutil.Blur(&pfutil.RealPartSlice{Data: field.Data}, []int{N, N}, &pfutil.BoxKernel{Width: 2})

	model := NewModel()
	model.AddField(field)
	model.AddScalar(NewScalar("gamma", complex(gamma, 0.0)))
	model.RegisterFunction("MINUS_CHEM_POT", func(i int, bricks map[string]Brick) complex128 {
		phi := bricks["phi"].Get(i)
		return (1.0 - phi*phi) * (phi + complex(3.0*rho/4.0, 0.0))
	})
	model.AddEquation("dphi/dt = MINUS_CHEM_POT + gamma*LAP phi")

	sdd := NewSDD([]int{N, N}, &model)
	sdd.TimeConstants.Orientation = 1.0
	sdd.TimeConstants.DimerLength = 1.0
	sdd.InitDimerLength = 1.0
	sdd.MinDimerLength = 5e-6
	dt := 0.7
	sdd.Dt = dt
	sdd.Init([]Field{init}, []Field{final})

	solver := NewSolver(&model, []int{N, N}, dt)
	solver.Stepper = &sdd

	fileIO := CsvIO{
		Prefix:     folder + "phi",
		DomainSize: []int{N, N},
	}

	nsteps := int(1000.0 / dt)

	if storeFiles {
		monitorLog, _ := os.Create(folder + "sddMonitor.csv")
		defer monitorLog.Close()
		sdd.Monitor.LogFile = monitorLog

		solver.AddCallback(func(s *Solver, epoch int) {
			sdd.SaveOrientation(fmt.Sprintf(folder+"orientation%d.csv", epoch))
		})
		solver.AddCallback(sdd.Monitor.Log)
		solver.AddCallback(fileIO.SaveFields)
	}

	solver.Solve(nsteps, 1)

	data := model.Fields[0].Data
	for i := range data {
		if cmplx.IsNaN(data[i]) {
			t.Errorf("NaN detected in solution\n%v\n", data)
			return
		}
	}

	// Calculate the area of the critial droplet
	area := 0.0
	for i := range data {
		area += 0.5 * (1.0 + real(data[i]))
	}
	Rc := math.Sqrt(area / math.Pi)

	if math.Abs(Rc-rc) > 0.3 {
		t.Errorf("Expected radius %f got %f\n", rc, Rc)
	}
}

func TestDiagonalShermannMorrison(t *testing.T) {
	A := mat.NewDense(3, 3, []float64{
		0.4, 0.0, 0.0,
		0.0, -0.2, 0.0,
		0.0, 0.0, 1.4,
	})

	u := []float64{-1.0, 2.0, 3.0}
	v := []float64{2.0, 2.3, 1.2}

	b := mat.NewVecDense(3, []float64{4.0, 6.0, 8.2})

	bCmplx := make([]complex128, 3)
	for i := 0; i < 3; i++ {
		bCmplx[i] = complex(b.AtVec(i), 0.0)
	}

	dsm := diagonalShermannMorrison{
		invDiagonal: make([]complex128, 3),
		u:           make([]complex128, 3),
		v:           make([]complex128, 3),
	}

	// Update the a matrix
	for i := 0; i < 3; i++ {
		dsm.invDiagonal[i] = complex(1.0/A.At(i, i), 0.0)
		dsm.u[i] = complex(u[i], 0.0)
		dsm.v[i] = complex(v[i], 0.0)
		for j := 0; j < 3; j++ {
			A.Set(i, j, A.At(i, j)+u[i]*v[j])
		}
	}

	res := mat.NewVecDense(3, nil)
	res.SolveVec(A, b)
	check := mat.NewVecDense(3, nil)
	check.MulVec(A, res)

	if !mat.EqualApprox(check, b, 1e-10) {
		t.Errorf("Gonum solve failed. Expected\n%v\nGot\n%v\n", mat.Formatted(b), mat.Formatted(check))
	}

	dsm.dot(bCmplx)

	for i := 0; i < 3; i++ {
		v1 := res.At(i, 0)
		v2 := real(bCmplx[i])
		if math.Abs(v1-v2) > 1e-10 {
			t.Errorf("Expected %f got %f\n", v1, v2)
		}
	}
}

func ExampleModel() (*Model, *Solver) {
	N := 16
	f1 := NewField("conc", N*N, nil)
	for i := range f1.Data {
		if i > 5 {
			f1.Data[i] = complex(0.1, 0.0)
		}
	}
	model := NewModel()
	model.AddField(f1)
	model.AddEquation("dconc/dt = conc^3 - conc + LAP conc")

	dt := 0.01
	solver := NewSolver(&model, []int{N, N}, dt)
	return &model, solver
}

func TestRevertOrientationVector(t *testing.T) {
	model, solver := ExampleModel()
	N := int(math.Sqrt(float64(len(model.Fields[0].Data))))

	stepper := NewSDD([]int{N, N}, model)
	orient := make([]float64, N*N)
	for i := range orient {
		if i > 5 {
			orient[i] = -1.0
		} else {
			orient[i] = 1.0
		}
	}
	stepper.SetInitialOrientation(orient)
	dt := 0.01
	stepper.Dt = dt
	solver.Stepper = &stepper

	origFields := make([]complex128, N*N)
	copy(origFields, model.Fields[0].Data)
	solver.Solve(100, 1)

	fieldData := make([]complex128, N*N)
	copy(fieldData, model.Fields[0].Data)

	copy(model.Fields[0].Data, origFields)

	stepper = NewSDD([]int{N, N}, model)
	for i := range orient {
		orient[i] *= -1.0
	}
	stepper.SetInitialOrientation(orient)
	stepper.Dt = dt

	solver.Solve(100, 1)
	if !pfutil.CmplxEqualApprox(model.Fields[0].Data, fieldData, 1e-10) {
		t.Errorf("Solution not invariant under sign reversion. Expected\n%v\nGot\n%v\n", fieldData, model.Fields[0].Data)
	}
}

func TestPanicOnZeroTimeStep(t *testing.T) {
	model, solver := ExampleModel()
	N := int(math.Sqrt(float64(model.NumNodes())))
	stepper := NewSDD([]int{N, N}, model)
	orient := make([]float64, N*N)
	for i := range orient {
		orient[i] = 1.0
	}
	stepper.SetInitialOrientation(orient)

	for i, dt := range []float64{0.0, 0.3} {
		stepper.Dt = dt
		solver.Stepper = &stepper
		func() {
			defer func() {
				if i == 0 && recover() == nil {
					t.Errorf("Should panic if timestep is zero.")
				} else if i == 1 && recover() != nil {
					t.Errorf("Should not panic when the timestep is larger than zero.")
				}
			}()
			solver.Solve(10, 1)
		}()
	}
}

func TestHouseHolder(t *testing.T) {
	N := 4
	model := NewModel()
	f1 := NewField("field", N*N, nil)
	model.AddField(f1)
	sdd := NewSDD([]int{N, N}, &model)

	vec := make([]complex128, N*N)
	orient := make([]complex128, N*N)
	for i := 0; i < N*N; i++ {
		vec[i] = complex(float64(i), 0.0)
		orient[i] = complex(float64(i*i-3*i), 0.0)
	}

	// Normalize orientation
	length := 0.0
	for i := range orient {
		length += real(orient[i]) * real(orient[i])
	}
	length = math.Sqrt(length)
	for i := range orient {
		orient[i] /= complex(length, 0.0)
	}

	sdd.ft.FFT(vec)
	sdd.ft.FFT(orient)
	sdd.householder(vec, orient, 2.0, len(vec))
	sdd.ft.IFFT(vec)
	pfutil.DivRealScalar(vec, float64(len(vec)))

	// Expected result from numpy
	expect := []float64{0.0, 1.41446756, 2.41446756, 3.0,
		3.17106489, 2.92766222, 2.26979199, 1.19745421,
		-0.28935113, -2.19062403, -4.50636448, -7.23657249,
		-10.38124806, -13.94039118, -17.91400186, -22.3020801}

	tol := 1e-6
	for i := range expect {
		if math.Abs(expect[i]-real(vec[i])) > tol {
			t.Errorf("Expected %f got %f\n", expect[i], real(vec[i]))
		}
	}
}

func TestHouseHolderDenum(t *testing.T) {
	// Apply the reflection to the equation
	// x_{n+1} = x_n + dt*(I - sigma*vv^T)x_{n+1}
	sigma := 1.0
	dt := 0.8
	model, _ := ExampleModel()
	N := int(math.Sqrt(float64(model.NumNodes())))

	sdd := NewSDD([]int{N, N}, model)
	sdd.Dt = dt

	A := mat.NewDense(N*N, N*N, nil)
	orient := make([]float64, N*N)
	x := make([]float64, N*N)
	length := 0.0
	for i := range orient {
		orient[i] = 0.1 * float64(i*i-4*i)
		x[i] = 0.1 * float64(i)
		length += orient[i] * orient[i]
	}
	for i := range orient {
		orient[i] /= math.Sqrt(length)
	}

	for i := 0; i < N*N; i++ {
		A.Set(i, i, 1.0-dt)
		for j := 0; j < N*N; j++ {
			A.Set(i, j, A.At(i, j)+dt*sigma*orient[i]*orient[j])
		}
	}

	res := mat.NewVecDense(N*N, nil)
	res.SolveVec(A, mat.NewVecDense(N*N, x))

	// Solve the equation using the Shermann-Morrison method implemented in SDD method
	cOrient := make([]complex128, len(orient))
	cOnes := make([]complex128, len(x))
	cX := make([]complex128, len(x))
	for i := range cOrient {
		cOrient[i] = complex(orient[i], 0.0)
		cX[i] = complex(x[i], 0.0)
		cOnes[i] = complex(1.0, 0.0)
	}

	sdd.ft.FFT(cOrient)
	sdd.ft.FFT(cX)
	dsm := sdd.householderDenum(cOnes, cOrient, sigma)
	dsm.dot(cX)
	sdd.ft.IFFT(cX)
	pfutil.DivRealScalar(cX, float64(len(cX)))

	// Compare with the result obtained in real space
	tol := 1e-10
	for i := range cX {
		if math.Abs(real(cX[i])-res.AtVec(i)) > tol {
			t.Errorf("Expected %f got %f\n", res.AtVec(i), real(cX[i]))
		}
	}
}
