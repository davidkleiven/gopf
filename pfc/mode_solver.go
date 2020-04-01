package pfc

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
)

// ModeSolver directly solves the phase field equations by expanding the density
// field into modes. The modes are assumed to correspond to peaks in the free energy
// spectrum
type ModeSolver struct {
	IdealMix IdealMix

	// Miller is a list of miller corresponding to each mode
	Miller []Miller

	// ReciprocalCell is a type that represents the reciprocal lattice
	ReciprocalCell ReciprocalCell
}

// ModeDensity returns the contribution to the total density from
// the mode being passed. The position should be given in scaled
// coordinates. The contribution to the total density from onde mode
// is given by sum_{hkl} cos(g_{hkl}*r), where the sum runs over all
// equivalent miller indices. r is the position vector.
func (ms *ModeSolver) ModeDensity(mode int, scaledPosition []float64) float64 {
	density := 0.0
	equiv := EquivalentMiller(ms.Miller[mode])
	dim := len(scaledPosition)
	for _, miller := range equiv {
		if dim == 2 && miller.L != 0 {
			// In 2D, L should always be zero
			continue
		}
		arg := float64(miller.H)*scaledPosition[0] + float64(miller.K)*scaledPosition[1]

		if dim == 3 {
			arg += float64(miller.L) * scaledPosition[2]
		}
		density += math.Cos(2.0 * math.Pi * arg)
	}
	return density
}

// UnitcellIntegral integrates powers of mode contributions over the
// unitcell. Powers is a an array of the powers of each term in the
// same order as the modes are listed in the Miller attribute. The order
// of the Gauss-Quadrature rule used internally is passed as order
//
// Example:
// If the density consists of three modes, let's say [[100], [111], [200]]
// denotet n_{100}, n_{111} and n_{200}, respectively. If power = [1, 2, 3]
// this corresponds to the integral
//       **
//  1   *
//----- * dV n_{100}^1 * n_{111}^2 * n_{200}^3
//  V   *
//    **unit cell
// where V is the volume of the unit cell
func (ms *ModeSolver) UnitcellIntegral(powers []int, order int) float64 {
	if len(powers) != len(ms.Miller) {
		panic("modesolver: The length of the powers array has to match the number of modes.")
	}
	dim, _ := ms.ReciprocalCell.CellVec.Dims()
	if dim == 2 {
		integrand := func(sx, sy float64) float64 {
			res := 1.0
			for mode, p := range powers {
				if p != 0 {
					res *= math.Pow(ms.ModeDensity(mode, []float64{sx, sy}), float64(p))
				}
			}
			return res
		}
		return QuadSquare(integrand, order)
	}

	// 3D case
	integrand := func(sx, sy, sz float64) float64 {
		res := 1.0
		for mode, p := range powers {
			if p != 0 {
				res *= math.Pow(ms.ModeDensity(mode, []float64{sx, sy, sz}), float64(p))
			}
		}
		return res
	}
	return QuadCube(integrand, order)
}

// freeEnergyTerms traces the powers of the different modes as well
// as the accumulated integrals
type freeEnergyTerm struct {
	ModePowers []int
	Coeff      float64
}

// Power returns the total power (sum of ModePowers)
func (fe *freeEnergyTerm) Power() int {
	p := 0
	for i := range fe.ModePowers {
		p += fe.ModePowers[i]
	}
	return p
}

// power2key creates a string representation of the power
// the key is constructed as follows m<mode nr>p<power>
// m0p2m1p3 means the the zeroth mode is raised to a power 2 and the
// first mode is raise to power 3
func power2key(powers []int) string {
	key := ""
	for i, p := range powers {
		if p != 0 {
			key += fmt.Sprintf("m%dp%d", i, p)
		}
	}
	return key
}

// build constructs all terms associated with raising the total
// density to a particular power. Order is the order of the quadrature
// rule used to solve the integrals. The function returns map with all the
// terms originating from raising the density to a power
//
// Example:
// If the density is given by n = A_0*n_0 + A_1*n_1 and power is 2
// the map contains the following terms
//
// m0p2: ModePowers: []{2, 0}, Coeff: integral of n_1^2
// m1p2: ModePowers: []{0, 2}, Coeff: integral of n_2^2
// m0p1m1p1: ModePowers: []{1, 1}, Coeff: 2*integral of n_1*n_2
func (ms *ModeSolver) build(power int, order int) map[string]freeEnergyTerm {
	terms := make(map[string]freeEnergyTerm)
	end := make([]int, power)
	for i := range end {
		end[i] = len(ms.Miller)
	}
	comb := NewProduct(end)

	for idx := comb.Next(); idx != nil; idx = comb.Next() {
		powers := make([]int, len(ms.Miller))
		for j := range idx {
			powers[idx[j]]++
		}

		integral := ms.UnitcellIntegral(powers, order)
		key := power2key(powers)

		if val, ok := terms[key]; ok {
			newTerm := freeEnergyTerm{
				ModePowers: powers,
				Coeff:      val.Coeff + integral,
			}
			terms[key] = newTerm
		} else {
			terms[key] = freeEnergyTerm{
				ModePowers: powers,
				Coeff:      integral,
			}
		}
	}
	return terms
}

// mergeTerms merges the terms from src into dst
func mergeTerms(dst map[string]freeEnergyTerm, src map[string]freeEnergyTerm) {
	for k, v := range src {
		dst[k] = v
	}
}

// ModeResult is a type that is returned by the mode optimizer
type ModeResult struct {
	Energy     float64
	Amplitudes []float64
}

// Solve returns the coefficients that minimizes the free energy.
// initGuess provised an explorer such that the method can try many different
// starting positions, and then return the best. If nil, the HyperOctantSearcher
// will be used.
func (ms *ModeSolver) Solve(initGuess InitialGuessExplorer) ModeResult {
	// Evaluate integrals arising from the ideal mixture part using
	// 16 point quadrature rule
	idealMixPower3 := ms.build(3, 32)
	idealMixPower4 := ms.build(4, 32)

	// Update integrals with the correct prefactor from the ideal mixture
	for k, v := range idealMixPower3 {
		tmp := v
		tmp.Coeff *= ms.IdealMix.ThirdOrderPrefactor()
		idealMixPower3[k] = tmp
	}
	for k, v := range idealMixPower4 {
		tmp := v
		tmp.Coeff *= ms.IdealMix.FourthOrderPrefactor()
		idealMixPower4[k] = tmp
	}

	// Merge the terms
	allTerms := make(map[string]freeEnergyTerm)
	mergeTerms(allTerms, idealMixPower3)
	mergeTerms(allTerms, idealMixPower4)
	terms := make(map[string]freeEnergyTerm)
	rtol := 1e-4
	maxcoeff := 0.0
	for _, v := range allTerms {
		if math.Abs(v.Coeff) > maxcoeff {
			maxcoeff = math.Abs(v.Coeff)
		}
	}

	for k, v := range allTerms {
		if math.Abs(v.Coeff) > rtol*maxcoeff {
			terms[k] = v
		}
	}

	// Definer the minimization problem
	problem := optimize.Problem{
		// Cost function
		Func: func(amp []float64) float64 {
			totEnergy := 0.0
			for _, term := range terms {
				ampProd := 1.0
				for j := range amp {
					ampProd *= math.Pow(amp[j], float64(term.ModePowers[j]))
				}
				totEnergy += ampProd * term.Coeff
			}
			return totEnergy
		},

		// Gradient with respect to the amplitudes
		Grad: func(grad, amp []float64) {
			for dir := range amp {
				grad[dir] = 0.0
				for _, term := range terms {
					if term.ModePowers[dir] == 0 {
						continue
					}
					p := float64(term.ModePowers[dir])
					ampProd := 1.0
					for j := range amp {
						if j == dir {
							ampProd *= p * math.Pow(amp[j], p-1)
						} else {
							ampProd *= math.Pow(amp[j], float64(term.ModePowers[j]))
						}
					}
					grad[dir] += ampProd * term.Coeff
				}
			}
		},

		// Hessian with respect to the amplitudes
		Hess: func(hess *mat.SymDense, amp []float64) {
			for dir1 := range amp {
				for dir2 := dir1; dir2 < len(amp); dir2++ {
					hess.SetSym(dir1, dir2, 0.0)
					value := 0.0
					for _, term := range terms {
						ampProd := 1.0
						for j := range amp {
							if j == dir1 && j == dir2 {
								p := float64(term.ModePowers[j])
								if term.ModePowers[j] < 2 {
									ampProd *= 0.0
								} else {
									ampProd *= p * (p - 1) * math.Pow(amp[j], p-2)
								}
							} else if j == dir1 || j == dir2 {
								if term.ModePowers[j] == 0 {
									ampProd *= 0.0
								} else {
									p := float64(term.ModePowers[j])
									ampProd *= p * math.Pow(amp[j], p-1)
								}
							} else {
								ampProd *= math.Pow(amp[j], float64(term.ModePowers[j]))
							}
						}
						value += ampProd * term.Coeff
					}
					hess.SetSym(dir1, dir2, value)
				}
			}
		},
	}

	// Minimize the sum of ideal mixing energy and excess energy
	if initGuess == nil {
		amplitudes := make([]float64, len(ms.Miller))
		for i := range amplitudes {
			amplitudes[i] = 0.1
		}
		initGuess = &HyperOctantExplorer{X0: amplitudes}
	}

	result := ModeResult{
		Energy:     0.0,
		Amplitudes: make([]float64, len(ms.Miller)),
	}
	for x0 := initGuess.Next(); x0 != nil; x0 = initGuess.Next() {
		opt, _ := optimize.Minimize(problem, x0, nil, &optimize.Newton{})

		if opt.Location.F < result.Energy {
			result.Energy = opt.Location.F
			copy(result.Amplitudes, opt.Location.X)
		}
		initGuess.OnMinimizationFinished(opt)
	}
	return result
}

// ElasticPropResult returns the obtained amplitide as well as the
// elastic tensor that corresponds to the solution
type ElasticPropResult struct {
	Amp []float64
	Fit map[string]float64
}

// ElasticTargetType is an integer type used to tune peak widths
type ElasticTargetType int

const (
	// ISOTROPIC is used together with ElasticPropertyTarget type to indicate
	// that the bulk and shear modulus should be fitted
	ISOTROPIC ElasticTargetType = iota
)

// ElasticPropertyTarget is a type that holds target values for fitting widths of
// peak heights. The default is to fit an isotropic material (in which case the
// bulk modules and shear modulus are matched)
type ElasticPropertyTarget struct {
	Bulk  float64
	Shear float64

	TargetType ElasticTargetType
}

// MatchElasticProp calculates a set of prefactors associated with each mode
// peak such that it matches the passed elastic tensor. The expressions are
// obtained as follows are obtained by taking second derivatives with respect
// to the strain tensor. From the n^2 term in the ideal mixture term and the
// excess energy we obtain the following relation for the contribution to the
// the total energy
// F_q = 0.5*beta_q*A_q*(1 - (1 - dk_q^2/2(w_q)))
// where V_cell is the volume of the unit cell, q is a multiindex representing a
// hkl tuple. beta_q is the number of equivalent planes, A_eq is the amplitude
// associated with the mode. Finally, the gaussian peak as been expanded around
// the maximum to leading order in dk_q, which is the change in mode location
// due to the strain. w_q is width of the peak. In the following we introduce
// a set of factors h_q^2 = 0.5*beta_q*A_q/w_q. The contribution to the free
// energy from mode q is therefore given by
// F_q = 0.5*h_q^2*dk^2
// The change in reciprocal lattice originating from a small strain is to first order
// dC = -C*e, where e is the strain tensor and C is the original reciprocal lattice
// A given hkl vector is given by dk_q = C*m_q, where m_q is a vector of miller indices.
// Further, the length of the dk_q square is given by dk_q^Tdk_q = m^T*e^T*C^T*C*e*m
// Differentiation twice with respect to an element in the strain tensor (let's say st and uv)
// we get that C_{stuv} = -h_q*m_t*Q_su*m_v, where Q = C^TC has been introduced.
// To obtain the total free energy, we sum over all modes q. This leaves, a linear
// system of equations to be solved. It is possible to fit the following properties
// 1. Bulk - bulk modulus in 3D
// 2. shear - shear modulus in 3D
// 3. area - area modulus (the equivalent of bulk modulus, but in 2D)
// 4. xyshear - shear modulus in 2D (e.g. shear in xy plane)
func (ms *ModeSolver) MatchElasticProp(properties map[string]float64) ElasticPropResult {
	weights := make(map[string]func(Miller, *mat.Dense) float64)
	weights["bulk"] = BulkModWeight
	weights["shear"] = ShearModWeight
	weights["area"] = AreaModulusWeight
	weights["xyshear"] = XYPlaneShearWeight

	// Check that we know how to calculate the requested properties
	for k := range properties {
		if _, ok := weights[k]; !ok {
			msg := fmt.Sprintf("Unknown property %s", k)
			panic(msg)
		}
	}
	penalty := 0.0
	if len(properties) < len(ms.Miller) {
		penalty = 1e-6 // Add a small regularization to the problem
	}
	r, c := ms.ReciprocalCell.CellVec.Dims()
	Q := mat.NewDense(r, c, nil)
	Q.Mul(ms.ReciprocalCell.CellVec.T(), ms.ReciprocalCell.CellVec)

	// Define the optimization problem
	problem := optimize.Problem{
		Func: func(x []float64) float64 {
			value := 0.0
			for k, v := range properties {
				pred := 0.0
				for i := range x {
					pred += x[i] * x[i] * weights[k](ms.Miller[i], Q)
				}
				value += math.Pow(v-pred, 2.0)
			}

			for i := range x {
				value += penalty * x[i] * x[i] * x[i] * x[i]
			}
			return value
		},
		Grad: func(grad []float64, x []float64) {
			for i := range grad {
				grad[i] = 0.0
			}
			for k, v := range properties {
				pred := 0.0
				for i := range x {
					pred += x[i] * x[i] * weights[k](ms.Miller[i], Q)
				}
				for j := range grad {
					grad[j] -= 4.0 * (v - pred) * x[j] * weights[k](ms.Miller[j], Q)
				}
			}

			for i := range x {
				grad[i] += 4.0 * penalty * x[i] * x[i] * x[i]
			}
		},
	}

	initX := make([]float64, len(ms.Miller))
	for i := range initX {
		initX[i] = 1.0
	}
	initX[0] = 1.0
	initX[1] = 0.01
	optRes, _ := optimize.Minimize(problem, initX, nil, nil)

	fitRes := ElasticPropResult{
		Amp: optRes.Location.X,
		Fit: make(map[string]float64),
	}

	for k := range properties {
		value := 0.0
		for i := range fitRes.Amp {
			value += fitRes.Amp[i] * fitRes.Amp[i] * weights[k](ms.Miller[i], Q)
		}
		fitRes.Fit[k] = value
	}
	return fitRes
}

// BulkModWeight returns the weight (e.g. value in design matrix)
// needed to fit a bulk modulus. Miller is the mode, Q is that matrix
// product C^TC, where each column in C represents a reciprocal lattice vector
// The Reuss formula for the bulk modulus is applied
// https://wiki.materialsproject.org/Elasticity_calculations
func BulkModWeight(miller Miller, Q *mat.Dense) float64 {
	weight := 0.0

	// Diagonal terms
	r, _ := Q.Dims()
	for i := 0; i < r; i++ {
		weight += float64(miller.At(i)) * Q.At(i, i) * float64(miller.At(i))
	}

	// Off diagonal terms
	for i := 0; i < r; i++ {
		for j := i + 1; j < r; j++ {
			weight += 2.0 * float64(miller.At(i)) * Q.At(i, j) * float64(miller.At(j))
		}
	}
	return weight / 9.0
}

// ShearModWeight calcualtes the weight needed to fit the shear modulus
// Miller is the mode, Q is that matrix
// product C^TC, where each column in C represents a reciprocal lattice vector
// The Reuss formula for the shear modulus is applied
// https://wiki.materialsproject.org/Elasticity_calculations
func ShearModWeight(m Miller, Q *mat.Dense) float64 {
	weight := 0.0
	// Diagonal terms
	dim, _ := Q.Dims()
	for i := 0; i < dim; i++ {
		weight += float64(m.At(i)) * Q.At(i, i) * float64(m.At(i))
	}

	// Off-diagonal terms
	for i := 0; i < dim; i++ {
		for j := i + 1; j < dim; j++ {
			weight -= float64(m.At(i)) * Q.At(i, j) * float64(m.At(j))
			weight += 3.0 * float64(m.At(j)) * Q.At(i, i) * float64(m.At(j))
		}
	}
	return weight / 15.0
}

// AreaModulusWeight return the weight associated with area compression
// this is the 2D equivalent of the bulk modulus. The modulus is obtained
// as follows. The area modulus is defined via B_A = A*dE^2/dA^2, where
// E is the total elastic energy
func AreaModulusWeight(miller Miller, Q *mat.Dense) float64 {
	m0 := float64(miller.At(0))
	m1 := float64(miller.At(1))
	weight := m0*Q.At(0, 0)*m0 + m1*Q.At(1, 1)*m1 + 2.0*m0*Q.At(0, 1)*m1
	return weight / 4.0
}

// XYPlaneShearWeight returns the shear modulus associated with shear in
// the XY plane. The values for the averages is obtained from
//
// Meille, S. and Garboczi, E.J., 2001.
// Linear elastic properties of 2D and 3D models of porous materials made
// from elongated objects.
// Modelling and Simulation in Materials Science and Engineering, 9(5), p.371.
func XYPlaneShearWeight(miller Miller, Q *mat.Dense) float64 {
	m0 := float64(miller.At(0))
	m1 := float64(miller.At(1))
	weight := m0*Q.At(0, 0)*m0 + m1*Q.At(1, 1)*m1 - 2.0*m0*Q.At(0, 1)*m1 + 4.0*m0*Q.At(1, 1)*m0
	return weight / 8.0
}

// PeakWidthResult a type that contains the widths of the correlation function peaks
// as well as the fitted elastic properties resulting from the fit
type PeakWidthResult struct {
	Widths []float64
	Fit    map[string]float64
}

// PeakWidths returns the width of the different peaks in the correlation function.
// The widths are obtained by matching elastic properties specified in target.
func (ms *ModeSolver) PeakWidths(target map[string]float64, modeAmp []float64) PeakWidthResult {
	ampResult := ms.MatchElasticProp(target)

	result := PeakWidthResult{
		Fit:    ampResult.Fit,
		Widths: make([]float64, len(modeAmp)),
	}

	dim, _ := ms.ReciprocalCell.CellVec.Dims()
	for i, a := range modeAmp {
		numEquiv := float64(NumEquivalent(ms.Miller[i], dim))
		result.Widths[i] = math.Sqrt(0.5*numEquiv) * math.Abs(a/ampResult.Amp[i])
	}
	return result
}
