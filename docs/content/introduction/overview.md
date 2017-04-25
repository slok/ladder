---
date: 2016-11-13T10:20:46Z
title: Overview
menu:
  main:
    parent: Introduction
    weight: 10

---

## What is Ladder for?

Ladder is a simple and small program that is used to autoscale *things*, when we
mean *things* we are talking about any kind of stuff that can increase or decrease,
for example, docker containers, AWS instances, simple binnary processes, files...

In a few words, Ladder can automate the process of increasing and decreasing the
quantity of objects based on inputs.

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
