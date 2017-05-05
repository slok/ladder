FROM themotion/ladder_base:latest

USER root

RUN apk --update add tar bash curl wget g++  && rm -rf /var/cache/apk/*

USER ladder

# Install dev dependency
RUN go get github.com/golang/mock/mockgen
RUN go get github.com/Masterminds/glide

# Github releaser
RUN go get github.com/aktau/github-release

WORKDIR /go/src/github.com/themotion/ladder/