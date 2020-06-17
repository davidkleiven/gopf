DBNAME="diffusion.db"

go run main.go

# Check the simulation ids
gopf db list $DBNAME -c simId

# Check all the comments
gopf db list $DBNAME -c comment -m 5

# Check creation time of the last entry
gopf db time $DBNAME

# Check last comment
gopf db comment $DBNAME

# Update last comment
gopf db comment $DBNAME -n "This is my new comment"
gopf db comment $DBNAME

# Export temperature to a csv file
gopf db export $DBNAME -t ts -o timeseries.csv

# Export field data
gopf db export $DBNAME -t fd -o concentration.csv

# Create a plot with the concentration
gopf contour -f concentration.csv -c conc -o concentration.png

# Show unique attribute names
gopf db attr $DBNAME --unique

# Show all attributes
gopf db attr $DBNAME

# Show a single attribute
gopf db attr $DBNAME -n meanConc

# Clean-up
rm timeseries.csv
rm concentration.csv
rm diffusion.db
rm concentration.png