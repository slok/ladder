---
date: 2016-11-13T10:47:00Z
title: Autoscaler

menu:
  main:
    parent: Concepts
    weight: 20
---

A ladder autoscaler is built with different kind of [blocks]({{< relref "concepts/blocks.md" >}}),
in the next section you can read the difference between these different kind of blocks.

## Flow

The flow of a Ladder autoscaler is very simple, at a high overview its as simple as this:

![Ladder flow](/img/flow.png)

## Pull based system

Ladder autoscalers are based on pull system, this means that an autoscaler will
check the state on regular intervals. Ladder philosophy is not to stay listening to some event
that triggers the scaling flow. Ladder phisolophy is to be automatic and autonomous

## Multiple autoscalers

Ladder runs multiple autoscalers at once, in different invervals based on the configuration,
for example a Ladder instance could be running these autoscalers:

* Cluster autoscaler that will check the state each 5 minutes
* Service X autoscaler taht willcheck the state each 1 minute
* Service Y autoscaler that will check the state each 10 minutes
* Size of a disk that will check the state each hour

These autoscalers are being check concurrently at different intervals by a single Ladder instance,
you can set any number of Autoscalers for each Ladder instance.

{{< note title="Note" >}}
You can set multiple Ladder instances with different autoscalers on each one
{{< /note >}}
