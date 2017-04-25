FROM golang:1.7-alpine


RUN apk --update add musl-dev gcc tar git bash wget && rm -rf /var/cache/apk/*

# Create user
ARG uid=1000
ARG gid=1000
RUN addgroup -g $gid ladder
RUN adduser -D -u $uid -G ladder ladder

RUN mkdir -p /go/src/github.com/themotion/ladder/
RUN chown -R ladder:ladder /go

WORKDIR /go/src/github.com/themotion/ladder/

USER ladder

# Install dependency manager
RUN go get github.com/Masterminds/glide
