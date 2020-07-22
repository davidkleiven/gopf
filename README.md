![Logo](docs/assets/logo.svg)

[![Build Status](https://travis-ci.org/davidkleiven/gopf.svg?branch=master)](https://travis-ci.org/davidkleiven/gopf)
[![Coverage Status](https://coveralls.io/repos/github/davidkleiven/gopf/badge.svg?branch=master)](https://coveralls.io/github/davidkleiven/gopf?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidkleiven/gopf)](https://goreportcard.com/report/github.com/davidkleiven/gopf)


Phase Field library in Go using a spectral solver. Go to the [webpage](https://davidkleiven.github.io/gopf/) to see examples.

# Features

* Spectral solver using [FFTW](https://github.com/barnex/fftw) to perform fourier transformes
* Supports 2D and 3D simulation domains
* Rich catalog with example applications. Some selected cases are explained in detail on our [webpage](https://davidkleiven.github.io/gopf/)
* Supports user defined terms/functions and equations
* Wide collection of ready-to-go terms (Khachaturyan elastic theory, white noise, volume conserving noise, Vandeven filters, pair correlation functions and many more)
* Flexible storage formats (SQL database, csv files, raw binary with XDMF)
* Command Line Interface (CLI) for interacting with the SQL database
* CLI for rapid production of contour plots of CSV files in case of 2D calculations

# Database For Phase Field Simulations

GOPF has support for storing collections of simulations in one SQLite database. This can be useful
if there are many simulations that are related (e.g. parameter sweeps). All simulations stored
in one DB must have the same domain size. The GOPF database supports

* Descriptive comments for each simulation (can changed from the CLI)
* Each simulation is given a timestamp when it starts
* Arbitrary numeric and/or string attributes
* Store all field data
* Store timeseries of user programmable function
* Export timeseries and field data to CSV files (useful if a third-party program is used for analysis)

For further information on the schema and example see the [database example](examples/database).