---
date: 2016-11-13T10:25:52Z
title: Installing
menu:
  main:
    parent: Introduction
    weight: 11
---

Ladder is written in Go, this makes it very portable, Go gives us the power
of sharing a single binary, easy and fast.


## Precompiled binary

TODO (when first version uploaded)

## Docker

An official Docker images is available:

```bash
$ docker pull themotion/ladder
$ docker run -p 9094:9094 themotion/ladder
```

## Building from source

You can build from source using go:

```bash
go get github.com/themotion/ladder
```

You can also build cloning the repository:

```bash
$ mkdir -p $GOPATH/src/github.com/themotion
$ cd $GOPATH/src/github.com/themotion
$ git clone https://github.com/themotion/ladder.git
$ cd ladder
$ make build_release
$ ./bin/ladder --help
```
