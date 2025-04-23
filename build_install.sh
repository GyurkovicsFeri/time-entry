#!/bin/bash

echo "Building time-entry..."

# Build the project
go build -o time-entry ./cli

echo "Installing time-entry..."

# Install the project
mv time-entry ~/go/bin/time-entry

echo "Done!"
