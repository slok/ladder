---
date: 2016-11-13T10:20:46Z
title: Overview
menu:
  main:
    parent: Introduction
    weight: 10

---

## What is Ladder for?

Ladder is a simple and flexible general purpose autoscaler.

The idea behind Ladder is to autoscale anything configuring and combining reusable [blocks](https://themotion.github.io/ladder/concepts/blocks/) 
of different types in a yaml file. These blocks are flexible and easy to extend, so anyone can use or develop any kind of scaling targets, 
policies or inputs.

Some examples that Ladder can do at this moment:

* Get number of messages in a [SQS queue](https://aws.amazon.com/sqs/), apply a constant factor to this input, then use this quantity to upscale or downscale the [EC2 machines](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Instances.html) of an [AWS AutoscalingGroup](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)
* Get the latency of a service from a [Prometheus](https://prometheus.io/) metric, if this latency is greater than 800ms, add one more instance to the actual number of instances of that service running on [ECS](https://aws.amazon.com/ecs/),
if is less than 200ms remove one instance to the running ones instead.

You want to start using it? Jump to the [tutorial]({{< relref "introduction/quickstart.md" >}})!

## What is not Ladder for?

Ladder is not for manual scaling, to ensure that an scaling target is always
at some quantity, or using Ladders metric to monitor the status of the target.
Although Ladder could do this, is not the main objective and depending on the
kind of target, there are other better tools for this purpouses out there.

## Features

* Very flexible and configurable
* Simple, light and fast
* Reliable
* Easy to configure and extend
* Metrics ready (Prometheus)
* Easy to deploy and set up running
* Tons of third party blocks ready to use (AWS, Prometheus...)

## Future

We want to add more blocks to the ones that Ladder provides by default (ECS & EC2 ASG), for example:

* Inputs:
    * Get metrics from Datadog
    * Get number of messages from Rabbitmq queue
* Filters:
    * Apply statistic prediction based on a metric, previous autoscaling result, etc
* Scalers:
    * Kubernetes replicas
    * Instance VMs on  GCE
    * Azure virtual machines