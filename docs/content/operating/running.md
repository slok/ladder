---
date: 2016-11-13T17:54:56Z
title: Running

menu:
  main:
    parent: Operating
    weight: 30
---

## Configuration file

When you run ladder by default will load `ladder.yml`, if not present you will
need to pass the configuration path to the conf file with the argument
`-config.file` or `--config.file`.

```bash
$ ladder -config.file=/etc/ladder/conf.yml
```

## Listen address

When running ladder by default will listen on `0.0.0.0:9094` but you can override
this using `-listen.address` or `--listen.address`

```bash
$ ladder -listen.address="127.0.0.1:9092"
```

## Debug

You can run ladder in debug mode using the `-debug` or `--debug` flag, this
will print debug messages and also register dummy blocks so you can use them
to fake inputs and arrangements

```bash
$ ladder -config.file=/etc/ladder/conf.yml -debug
```

## Dry run

You can run ladder in dy run mode using the `-dry.run` or `--dry.run` flag, this
will gather and arrange as a regular run but will omit the final step of
scaling up or down wo you can test your logic before deplying 100000 machines
and closing your company because of bankrupt


```bash
$ ladder -config.file=/etc/ladder/conf.yml -dry.run
```

## Json logger

By default ladder will log in text format, but in production when you use systems like
Elastic search, json is better because of the indexing, for this Ladder can log in
json just running with `-json.log` or `--json.log` flag.

```bash
$ ladder -config.file=/etc/ladder/conf.yml -json.log
```
