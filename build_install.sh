#!/bin/bash

# Build the project
go build -o time-entry ./cli

# Install the project
mv time-entry ~/go/bin/time-entry
