# Cahn-Hilliard Equation

Here, we show how the Cahn-Hilliard equation can be solved using GOPF.

<p align="center">
    <img src="figs/cahnHillOrig.png">
</p>

*M* is a mobility factor, gamma is a the gradient coefficient which is related to the surface tension and control the width of the diffuse interface.
We start the calculation by initializing the concentration field at random. Four snapshots of the concentration field at different times is shown below
<p>
 <img src="figs/conc0.png" width="400">
 <img src="figs/conc1.png" width="400">
</p>
<p>
 <img src="figs/conc2.png" width="400">
 <img src="figs/conc3.png" width="400">
</p>
We observe that a spinodal decomposition takes place, and gradually larger domains form.
