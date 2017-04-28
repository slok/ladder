#!/bin/bash -e

# Run vet
make vet

# Run tests
make test

# Push to docker hub only when master branch build
if [ $TRAVIS_BRANCH == 'master' ]; then
    make push
fi