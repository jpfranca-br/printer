#!/bin/bash

# Output binary directory
OUTPUT_DIR="build"
mkdir -p $OUTPUT_DIR

# Get all GOOS/GOARCH combinations
combinations=$(go tool dist list)

# Iterate through each combination
for combo in $combinations; do
    # Split the combination into GOOS and GOARCH
    GOOS=${combo%/*}
    GOARCH=${combo#*/}
    # Output file name
    OUTPUT_FILE="$OUTPUT_DIR/printer_${GOOS}_${GOARCH}"

    # Add .exe extension for Windows
    if [ "$GOOS" == "windows" ]; then
        OUTPUT_FILE="$OUTPUT_FILE.exe"
    fi

    # Build the binary
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -o $OUTPUT_FILE

    # Check if build succeeded
    if [ $? -ne 0 ]; then
        echo "Failed to build for $GOOS/$GOARCH"
    fi
done

echo "Builds are saved in $OUTPUT_DIR"
