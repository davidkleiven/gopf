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