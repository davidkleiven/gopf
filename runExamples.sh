echo "Running diffusion example"
go run examples/diffusion/main.go
rm *.bin
rm diffusionMonitor.json
rm diffusion.xdmf

echo "Running Cahn-Hilliard example"
go run examples/cahnHilliard/main.go
rm *.bin