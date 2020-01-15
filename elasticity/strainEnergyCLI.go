package elasticity

import (
	"math"

	"github.com/davidkleiven/gosfft/sfft"
	"gonum.org/v1/gonum/mat"
)

// StrainEnergyInput is a struct that holds all parameters needed
// to calculate the strain energy
type StrainEnergyInput struct {
	HalfA             float64
	HalfB             float64
	HalfC             float64
	Misfit            []float64
	MatPropMatrix     []float64
	MatPropInc        []float64
	DomainSize        int
	ApplyPerturbation bool
}

// CalculateStrainEnergy calculates the strain energy according to the parameters
// given in he passed parameters
func CalculateStrainEnergy(params StrainEnergyInput) float64 {
	matPropMatrix := FromFlatVoigt(params.MatPropMatrix)
	misfit := mat.NewDense(3, 3, nil)
	for i := 0; i < 3; i++ {
		misfit.Set(i, i, params.Misfit[i])
	}
	misfit.Set(0, 1, 0.5*params.Misfit[5])
	misfit.Set(1, 0, 0.5*params.Misfit[5])
	misfit.Set(0, 2, 0.5*params.Misfit[4])
	misfit.Set(2, 0, 0.5*params.Misfit[4])
	misfit.Set(1, 2, 0.5*params.Misfit[3])
	misfit.Set(2, 1, 0.5*params.Misfit[3])
	effForce := NewEffectiveForceFromMisfit(matPropMatrix, misfit)

	indicator := sfft.NewCMat3(params.DomainSize, params.DomainSize, params.DomainSize, nil)
	for i := 0; i < params.DomainSize; i++ {
		for j := 0; j < params.DomainSize; j++ {
			for k := 0; k < params.DomainSize; k++ {
				x := float64(i - params.DomainSize/2)
				y := float64(j - params.DomainSize/2)
				z := float64(k - params.DomainSize/2)
				v := math.Pow(x/params.HalfA, 2) + math.Pow(y/params.HalfB, 2) + math.Pow(z/params.HalfC, 2)

				if v <= 1.0 {
					indicator.Set(i, j, k, complex(1.0, 0.0))
				}
			}
		}
	}

	ft := sfft.NewFFT3(params.DomainSize, params.DomainSize, params.DomainSize)
	ft.FFT(indicator.Data)
	force := make([][]complex128, params.DomainSize*params.DomainSize*params.DomainSize)
	for k := range force {
		force[k] = make([]complex128, 3)
	}

	for comp := 0; comp < 3; comp++ {
		fComp := effForce.Get(comp, ft.Freq, indicator.Data)
		for k := range force {
			force[k][comp] = fComp[k]
		}
	}

	disp := Displacements(force, ft.Freq, &matPropMatrix)
	// Inservse FFT such that we can use it distinguish regions
	ft.IFFT(indicator.Data)
	for i := range indicator.Data {
		indicator.Data[i] /= complex(float64(len(indicator.Data)), 0.0)
	}

	diffMatProp := NewRank4()
	if params.MatPropInc != nil {
		propInc := FromFlatVoigt(params.MatPropInc)
		for i := range diffMatProp.Data {
			diffMatProp.Data[i] = propInc.Data[i] - matPropMatrix.Data[i]
		}

		if params.ApplyPerturbation {
			correctionForce := PerturbedForce(ft, misfit, disp, func(i int) float64 { return real(indicator.Data[i]) }, &diffMatProp)
			dispCorr := Displacements(correctionForce, ft.Freq, &matPropMatrix)

			// Add the correction to the original displacements
			for i := range dispCorr {
				for comp := 0; comp < 3; comp++ {
					disp[i][comp] += dispCorr[i][comp]
				}
			}
		}
	}

	energy := 0.0
	strains := make([]*mat.Dense, len(disp))
	for k := range strains {
		strains[k] = mat.NewDense(3, 3, nil)
	}

	for i := 0; i < 3; i++ {
		for j := i; j < 3; j++ {
			strain := Strain(disp, ft.Freq, i, j)
			ft.IFFT(strain)
			for k := range strain {
				re := real(strain[k]) / float64(len(strain))

				if real(indicator.Data[k]) > 0.5 {
					re -= misfit.At(i, j)
				}
				strains[k].Set(i, j, re)
				strains[k].Set(j, i, re)
			}
		}
	}
	for i := range strains {
		energy += EnergyDensity(&matPropMatrix, strains[i])
	}

	if params.MatPropMatrix != nil {
		for i := range strains {
			energy += EnergyDensity(&diffMatProp, strains[i])
		}
	}
	vol := 4.0 * math.Pi * params.HalfA * params.HalfB * params.HalfC / 3.0
	return energy / vol
}
