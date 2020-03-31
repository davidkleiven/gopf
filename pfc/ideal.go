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
// F_id = -(n^2/2 - C3*n^3/6 + C4*n^4/12)
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
	return 0.5*n*n - im.C3*n*n*n/6.0 + im.C4*n*n*n*n/12.0
}
