#!/usr/bin/env bash

TAG=$(git describe --tags --exact-match)

# Only release when tag
if ! [ -z ${TAG} ]; then
    # Check if already exists
    if $(github-release info -u themotion -r ladder | grep -q "${TAG},"); then
        echo "Release already exists"
        exit 0
    fi

    # Get if alpha release or release
    if [ -z $(echo ${TAG} | grep rc) ]; then
        # Get changelog, description and title for the realease
        CHANGELOG=$(sed  -n "/^## ${TAG}/,/^##/p"  ./CHANGELOG.md | head -n -1)
        DESC=$(echo -e "${CHANGELOG}" | tail -n +2)
        TITLE=$(echo -e "${CHANGELOG}" | head -1 | sed s/##\ //g)
    else
        TITLE="${TAG}"
        DESC="This is a pre release, please check the [changelog](CHANGELOG.md)"
        PRERELEASE="--pre-release"
    fi

    echo -e "Creating $TITLE release"
    github-release release \
        --user themotion \
        --repo ladder \
        --tag "${TAG}" \
        --name "${TITLE}" \
        --description "${DESC}" \
        ${PRERELEASE}
    
    # Build binary
    echo -e "Building Ladder"
    ./build.sh

    # Upload binary
    echo -e "Uploading Ladder binary"
    github-release upload \
    --user themotion \
    --repo ladder \
    --tag ${TAG} \
    --name "ladder-${TAG}.linux-amd64.bin" \
    --file bin/ladder
else
    echo "No tags no release, deal with it"
fi

