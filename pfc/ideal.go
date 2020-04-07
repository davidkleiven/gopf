package pfc

// IdealMix is a type used to represent the ideal mixing entropy with respect to a reference
// state. The polynomial is used together with a dimensionless density field n = rho/rho_0 - 1,
// where rho is the real density field and rho_0 is the reference density. The real mixing
// free energy is given by
//
// F_id = (1 + n)*ln(1 + n) - n
//
// An expansion of this around the reference state (n = 0) yields
//
// F_id = (n^2/2 - C3*n^3/6 + C4*n^4/12)
//
// where C3 and C4 are treated as free parameters that can be used to stabilize particular
// structures of interest. Reference
//
// Greenwood, M., Provatas, N. and Rottler, J., 2010.
// Free energy functionals for efficient phase field crystal modeling of structural phase transformations.
// Physical review letters, 105(4), p.045702.
type IdealMix struct {
	C3 float64
	C4 float64
}

// Eval evaluates the mixing entropy at a given dimensionless density
func (im *IdealMix) Eval(n float64) float64 {
	return im.QuadraticPrefactor()*n*n + im.ThirdOrderPrefactor()*n*n*n + im.FourthOrderPrefactor()*n*n*n*n
}

// Deriv returns the derivative with respect to the density
func (im *IdealMix) Deriv(n float64) float64 {
	return 2.0*im.QuadraticPrefactor()*n + 3.0*im.ThirdOrderPrefactor()*n*n + 4.0*im.FourthOrderPrefactor()*n*n*n
}

// QuadraticPrefactor returns the prefactor in front of the quadratic term
func (im *IdealMix) QuadraticPrefactor() float64 {
	return 0.5
}

// ThirdOrderPrefactor returns the prefactor in front of the third
// order term
func (im *IdealMix) ThirdOrderPrefactor() float64 {
	return -im.C3 / 6.0
}

// FourthOrderPrefactor returns the prefactor in front of the fourth
// order term
func (im *IdealMix) FourthOrderPrefactor() float64 {
	return im.C4 / 12.0
}
