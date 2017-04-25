---
date: 2016-11-13T15:05:48Z
title: Filters
menu:
  main:
    parent: Available blocks
    weight: 23
---

## Dummy

Dummy filterer will return a concrete value

### Name

`dummy`

### Options

```yaml
filters:
  - kind: dummy
    config:
```

{{< warning title="Warning" >}}Only used for testing{{< /warning >}}

## Limit

Limit filterer will restrict the scalign value to the max and min configured limits

### Name

`limit`

### Options

* `max`: The max value
* `min`: The min value

### Example

```yaml
filters:
  - kind: limit
    config:
      max: 10
      min: 2
```

## ECS running tasks

ECS running tasks filter will check how many tasks are pending by the ECS scheduler on a given
cluster, if this number exceeds the desired one then it will break the filters chain and set
the autoscaling to the current number.

### Name

`ecs_running_tasks`

### Options
* `aws_region`: String that contains the target ECS cluster region
* `cluster_name`: String that contains the name of the target ECS cluster
* `max_pending_tasks_allowed`: integer that describes the maximum number of allowed not running tasks (or pending tasks)
* `max_checks`: Number of failed checks failing before continuing with a regular scalation besides the last check
(if max checks is 0 then its disabled)
* `error_on_max_checks`: Boolean, when max checks is triggered instead of scaling as a regular iteration besides of the
    of the result it will return an error and stop this current iteration
* `when`: string (enum) can be `always`, `scale_up` or `scale_down` this will say when the filter will be applied

```yaml
filters:
  - kind: ecs_running_tasks
    config:
      aws_region: us-west-2
      cluster_name: slok-ECSCluster1-15OBYPKBNXIO6
      max_pending_tasks_allowed: 2
      max_checks: 5
      error_on_max_checks: false
      when: scale_down
```

## Scaling kind interval

scaling kind iternval will allow or not scaling if the scaling mode has been
active for the configured duration. for example if you want to scaleup after 30 seconds
The autoscaler will need to be in scaleup for more than 30 seconds without interruption
(this is changing scaling mode), this allows us to omit spikes for the interval
we want

### Name

`scaling_kind_interval`

### Options

* `scale_up_duration`: The duration of the scaling up mode need for triggering
* `scale_down_duration`: The duration of the scaling down mode need for triggering

### Example

```yaml
filters:
  - kind: scaling_kind_interval
    config:
      scale_up_duration: 30s
      scale_down_duration: 1m

```
