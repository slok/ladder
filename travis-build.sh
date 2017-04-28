#!/bin/bash -e

# Run vet
make vet

# Run tests
make test

# Push to docker hub only when master branch build or release
if [ ${TRAVIS_BRANCH} == 'master' ] || ! [ -z ${TRAVIS_TAG} ]; then
    make push
fi