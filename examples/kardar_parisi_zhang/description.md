# Kardar-Parisi-Zhang Equation

The Kardar-Parisi-Zhang (KPZ) equation non-linear stochastic partial differential equation. It describes the
height field *h* with respect to the time *t* and spatial coordinates *x*. The equation is given by

<p align="center"><img src="fig/kpz_eq.png"/></p>

where *&lambda;* and *&nu;* are tunable parameters. *&eta;* is a gaussian random noise term satisfying

<p align="center"><img src="fig/expectation.png"></p>

<p align="center"><img src="fig/autocorr.png"></p>

Snapshots of a soluton a *128x128* grid are shown below

<p align="center">
    <img src="fig/kpz0.png" width="400">
    <img src="fig/kpz1.png" width="400">
</p>
<p align="center">
    <img src="fig/kpz2.png" width="400">
    <img src="fig/kpz3.png" width="400">
</p>

An animation of the solution can be found on [youtube](https://www.youtube.com/watch?v=DlZTG_lcu90&feature=youtu.be). 
The average width *W* defined by

<p align="center"><img src="fig/width.png"></p>

where *A* is the total area, develops as shown below

<p align="center"><img src="fig/average_width.png">

Here, the dynamic exponent (*W(t) &prop; t<sup>&beta;</sup>*)is *&beta; = 0.235 &plusmn; 0.006* from powerlaw fits to 100 independent runs. Considering the
simpliclity of this setup (*128x128* grid), this is remarkably close to state-of-the-art estimates *&beta; = 0.2415 &plusmn; 0.0015* [1].

[1] [Kelling, Jeffrey, and Géza Ódor. "Extremely large-scale simulation of a Kardar-Parisi-Zhang model using graphics cards." Physical Review E 84.6 (2011): 061150.](https://doi.org/10.1103/PhysRevE.84.061150)