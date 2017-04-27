#!/bin/bash -e

# Run vet
make vet

# Run tests
make test

# Push to docker hub
make push