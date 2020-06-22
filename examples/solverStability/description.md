# Solver Stability
This examples shows the stability of different time-stepping schemes. As a test case we use
the Allan-Cahn equation

<p align="center">
    <img src="fig/allan_cahn.png">
</p>

here we use &gamma; = 0.02. We run 100 steps and checks if the solution goes to NaN.

Scheme\Timestep | 0.1 | 1.0 | 1.5 | 1.9 | 2.1 | 3.4 | 5.4 | 10.0 | 20.0 |
| ------------- | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-:  | :-:  |
euler | x | x | - | - | - | - | - | - | - |
rk4   | x | x | x | x | x | - | - | - | - |
Implicit Euler | x | x | x | x | x | x | x | x | x |

The RK4 scheme can be applied with almost two times larger time step. But it should be noted that
the computational cost on each scheme is higher than for the euler scheme. The implicit euler
is stable at even higher time steps. It is tested that it does converge for *dt=3.4* for this equation, but above that the non-linear solver fails to converge.