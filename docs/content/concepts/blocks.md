---
date: 2016-11-13T10:47:06Z
title: Blocks

menu:
  main:
    parent: Concepts
    weight: 21
---

As we said previously, Ladder can have multiple autoscalers, and each autoscaler is
made by differet kind of blocks, these blocks are:

* **Inputters**: The ones that gather data and make a decision
* **Solvers**: The ones that take multiple inputters result and take one of them
* **Filters**: The ones that take the Solver result and apply some kind of logic based on the solver result
* **Scalers**: The ones that scale based on all the previous flow result

Let take a look at Ladder's architecture first and then try to explain each of the parts


![Ladder architecture](/img/architecture.png)

## Inputter

An inputter is a a composite block, is made of a gatherer and an arranger.


{{< note title="Note" >}}
An autoscaler can have multiple inputters.
{{< /note >}}


## Gatherer

This kind of block is the one that will grab the values form external sources, is the
start of the autoscaler cycle.

Some kind of gatherers could be:

* The number of messages in a queue
* The CPU % used
* The mean request latency
* The remaining free size of a disk
* The temperature of a room

## Arranger

The arranger block is the one that makes the decisions based on the gatherer data. In order
to make a decision and return a valid quantity for the scaler block, the arranger has the gatherer data
and the current quantity of the scaler (or scaler target).

In other words, will take the gatherer data, the current scaler target data, apply some logic and return a result
that can be used by the scaler to set the arranged value on the scaler target.

Some kind of arrangers could be:

* If the current input is greater that a quantity during X minutes then increment by 10% the current scaler target quantity
* If the input is equal to a quantity then set the quantity to 0
* Take the current input data and add 5
* Take the current input data and divide by a constant value.


{{< note title="Note" >}}
The arranger of an inputter can be empty (this means that the gathered value will be passed transparently to the solver)
{{< /note >}}

## Solver

As we said above, we can have multiple inputters, but at the end the autoscaler only can set one correct value on the
scaling target, the solver is the one that takes all this inputters result and apply some logic so it finally returns
one single quantity result.

Some kind of solvers could be:

* The summatory of all inputters
* The max or min of all inputters
* The first of all the inputters
* A random of all the inputters

{{< note title="Note" >}}
If only one inputter is configured, the solver will be ignored
{{< /note >}}


## Filterer (filter)

Filterers or filters is a kind of block that is configured as an ordered list or chain.
Each filter will be applied in order and can change the input of the next filter to apply,
When executing a filter it will receive the arranged value by the solver and the current value of
the scaling target, after aplyign the filter logic, this will return a result, and a flag
specifying if the filter breaks the whole filter chain or not.

With this process we can validate, change or error multiple times the received value, before
the scaler scales the target

Some kind of filters could be:


* A limiting filterer that doesn't allow values more or lesser than the defined values
* A filter that breaks the chain without altering the current value if a check with a 3rd party service doesn't return correctly
* A interval filterer that allows scaling if the autoscaler has been in scaleup mode after 1 minute and allow downscale if the mode has been in downscale mode for 5 minutes


{{< note title="Note" >}}
The order of the filterers chain is very important
{{< /note >}}


## Scaler

An scaler is the final part of an autoscaler, this kind of blokc is the one that will trigger the scaling on the
scaling target with the received quantity (or if it didn't change then don't scale).

Some kind of scaler could be:

* The desired number of instances in an AWS auto scaling group
* The number of GBs asigned to an instance
* The asigned Memory to a docker container
* The number of replicas of a Kubernetes pod
