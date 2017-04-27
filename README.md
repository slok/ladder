# Ladder [![Build Status](https://travis-ci.org/themotion/ladder.svg?branch=master)](https://travis-ci.org/themotion/ladder)

Ladder is a simple and small program that is used to autoscale *things*, when we
mean *things* we are talking about any kind of stuff that can increase or decrease,
for example, docker containers, AWS instances, simple binnary processes, files...

## Features

* Very flexible and configurable
* Simple, light and fast
* Reliable
* Easy to configure and extend
* Metrics ready (Prometheus)
* Easy to deploy and set up running
* Tons of third party blocks ready to use (AWS, Prometheus...)

## Architecture

![](docs/static/img/architecture.png)


* Inputter: Gets data and returns an scaling quantity result
    * [Gatherer](https://themotion.github.io/ladder/blocks/gatherers/): Gets the data from any resource
    * [Arranger](https://themotion.github.io/ladder/blocks/arrangers/): Applies logic to convert the input to an scaling quantity result
* [Solver](https://themotion.github.io/ladder/blocks/solvers/): Takes one of the multiple results the inputters return.
* [Filter](https://themotion.github.io/ladder/blocks/filters/): Applies logic and changes(or not) the result of the Solver
* [Scaler](https://themotion.github.io/ladder/blocks/scalers/): Scales on a target the desired quantity received from the last filter applied, or from the solver if no filters where applied.


## Status

Ladder has been autoscaling TheMotion platform in production for more than 6 months. We need to finish
the documentantion like a quickstart or a tutorial, make the official Docker images and set a list of
supported providers (at this moment EC2 & ECS only) among other things.

## Documentation

Check out the online documentation at https://themotion.github.io/ladder or offline:

```bash
$ make serve_docs
```

Go to http://127.0.0.1:1313 on the browser

## Prometheus metrics

![](docs/static/img/grafana.png)

## Changelog

See [changelog](CHANGELOG.md)

## License

See [license](LICENSE)


## Authors

See [authors](AUTHORS) & [contributors](CONTRIBUTORS)

## Maintainers
See [maintainers](MAINTAINERS.md) to know who is/are the person/people you need to contact.