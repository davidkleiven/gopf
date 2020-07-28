echo "Running diffusion example"
go run examples/diffusion/main.go
rm *.bin
rm diffusionMonitor.json
rm diffusion.xdmf

echo "Running Cahn-Hilliard example"
go run examples/cahnHilliard/main.go
rm *.bin

echo "Running strain single precipitate example"
go run examples/strain_single_precipitate/main.go
rm *.bin
rm *.xdmf

echo "Running elasticity CLI"
go run cmd/gopf-elast-input/main.go -out="inputParams.json"
go run cmd/gopf-elast-energy/main.go -input="inputParams.json"
rm inputParams.json

echo "Running Kardar-Parisi-Zhang example"
go run examples/kardar_parisi_zhang/main.go -prefix="kpz"
rm *.bin
rm *.xdmf

echo "Running stepper stability"
go run examples/solverStability/main.go

echo "Running KKS example"
go run examples/kks/main.go
rm *.bin
rm *.xdmf

echo "Running build crystal example"
go run examples/buildCrystals/main.go
rm single_crystal.csv
rm grainBoundary.csv

echo "Running database CLI tests"
cd examples/database/
bash runCmds.sh
cd ../../

echo "Running charge transport example"
go run examples/electricConductivity/main.go
rm grainBoundaryCurrent.db
rm current.csv

echo "Running heterogeneous nucleation example"
go run examples/heterogeneousNucleation/main.go -dim 64 -innerR 0.05 -outerR 0.02 -dt 0.01 -epoch 2 -step 10

echo "Running shape example"
go run examples/drawing/main.go
rm shapes.csv