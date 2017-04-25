FROM themotion/ladder_base:latest
USER root
RUN apk add --no-cache g++

USER ladder
RUN go get github.com/golang/mock/mockgen
