---
date: 2017-04-29T09:18:33Z
title: Quickstart
menu:
  main:
    parent: Introduction
    weight: 12

---

Start making autoscalers in Ladder is very easy and fast, but if you haven't crafted any before
you may feel that you don't know where to start, this small tutorials will guide you so you understard
where to start and make your first autoscalers.

All the tutorials will use the [official Docker image](https://hub.docker.com/r/themotion/ladder/tags/)

## Basic tutorial

For a basic tutorial we will keep it simple, a small autoscaler that will set machines on a [AWS ASG](http://docs.aws.amazon.com/autoscaling/latest/userguide/AutoScalingGroup.html)
based on the number of messages on a [AWS SQS](https://aws.amazon.com/sqs/) and aplying a constant factor.

### Global configuration file

The first thing is to create `ladder.yml` file, this file has Ladder's global configuration:

```yaml
global:
  warmup: 30s

autoscaler_files:
  - "/etc/ladder/cfg-autoscalers/tutorial1/*.yml"
```

We are using default settings except the warmup, warmup setting will make that any of the autoscaler doesn't do anything until that duration has passed.

Also we point to the autoscaler configuration, in this case all the files that are on `tutorial1`.

### Autoscaler configuration file

Now we will create our autoscaler file, note that you may have multiple autoscaler files, and also
multiple autoscalers per file, but for this example we'll do it simple and use a single file and one
autoscaler.

We will start creating the file block by block.

### Autoscaler global settings

```yaml
name: basic_tutorial_autoscaler_sqs_asg
description: >
    basic tutorial autoscaler will set an AWS autoscaling group quantity based on the number of
    messages that an AWS SQS queue has at a given point aplying a constant factor
interval: 2m
scaling_wait_timeout: 3m
```

Our autoscaler is called `basic_tutorial_autoscaler_sqs_asg` and will run every `2m` to check if it needs to upscale or downscale, as a security measure we apply a `3m` timeout to the waiting time of the scaler when it decides to scale (an scaler waits until the scalation that has been applied has completed).

### Autoscaler input settings

In this case our autoscaler will have only one inputter to keep it simple. The inputter will return
the number of machines that wants based on an input, in our case the number of messages in a queue

```yaml
inputters:
  - name: sqs_queue_basic_tutorial_job_number
    description: "Get quantity based on the messages on the queue aplying a constant factor"
    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.eu-west-1.amazonaws.com/xxxxxxxxxxx/basic-tutorial-queue"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "eu-west-1"
    arrange:
      kind: constant_factor
      config:
        factor: 20
        round_type: "ceil"
```

This inputter has 2 subblocks, a [gantherer]({{< relref "blocks/gatherers.md" >}}) the one that gets the number of messages from the queue, and
an [arranger]({{< relref "blocks/arrangers.md" >}}) the one that applies the logic to convert the number on messages quantity to number of machines
desired quantity.

In this case we will get the number of messages from the queue (in AWS SQS is known as `ApproximateNumberOfMessages`) and will apply to this number a constant factor of 20(simple division)
always rounding up. For example:

| messages  | machines |
|---------- | -------- |
| 20        | 2        |
| 50        | 3        |
| 100       | 5        |
| 5500      | 275      |
| 54321     | 2717     |
| 100000    | 5000     |

### Autoscaler solver settings

In this case as we only have on inputter the [solver]({{< relref "blocks/solvers.md" >}}) block is not neccessary because we don't have
the need to decide wich of the inputters will be the chosen one.

#### Autoscaler filters settings

We will apply 2 [filters]({{< relref "blocks/filters.md" >}}) as a filter chain. They will take the result returned by the inputter and apply
the filters as a pipeline, first filter output will be the input of the next one and so on.

```yaml
filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 2m
        scale_down_duration: 10m

    - kind: limit
      config:
        max: 1000
        min: 2
```

Filters will be applied in the same order as defined in the configuration file, in this case first will
be applied the `scaling_kind_interval` filter and then `limit`.

We set `scaling_kind_interval`so we don't have spikes of machines, this filter will wait a concrete duration
before letting us to scale up or down, in this case we will wait at least for `2m` before scaling up (this can
be multiple iterations  of the autoscaler without doing nothing until it satisfies `2m` in scale up mode). On the
other hand we have `10m` to scale down, the same logic applies, it needs to be in scaling down mode for `10m` at least. As you can see we scale up fast, but scale down slowly, this is a common practice so we use the scaled up power
in case a new spike comes, and also we don't disturb our platform downscaling too fast.

The next filter is the safety filter that **always** should be set, is optional but you should set a limit.
In our case we want at least `2` machines running always, and tops we want `1000` (we don't want our company die due to an AWS invoice!)

### Autoscaler scaler settings

The [scaler]({{< relref "blocks/scalers.md" >}}) is the one that sets the quantity on the target, in our case an AWS ASG.

```yaml
scale:
    kind: aws_autoscaling_group
    config:
      auto_scaling_group_name: "basic-tutorial-AMIAutoScalingGroup"
      aws_region: "eu-west-1"
      scale_up_wait_duration: 1m
      scale_down_wait_duration: 5s
```

Our scaler will set `basic-tutorial-AMIAutoScalingGroup` ASG as its target, it will wait `1m` after scaling up before continuing with the next iteration and will wait `5s` before downscaling, in other words, it waits
machines to spin up (not too much) and it doesn't wait after start downscaling.

{{< note title="Note" >}}
ASG scaler will do its best so only scales down the machines next to the `1h` of utilization (by default, it will wait every machine to be running for `50m` at least), this is because AWS charges you per hour. ASG autoscaler wants to save you money!
{{< /note >}}

### Putting all together

`ladder.yml`

```yaml
global:
  warmup: 30s

autoscaler_files:
  - "/etc/ladder/cfg-autoscalers/tutorial1/*.yml"
```

`basic_tutorial.yaml`

```yaml
autoscalers:
- name: basic_tutorial_autoscaler_sqs_asg
  description: >
    basic tutorial autoscaler will set an AWS autoscaling group quantity based on the number of
    messages that an AWS SQS queue has at a given point aplying a constant factor
  interval: 2m
  scaling_wait_timeout: 3m

  scale:
    kind: aws_autoscaling_group
    config:
      auto_scaling_group_name: "basic-tutorial-AMIAutoScalingGroup"
      aws_region: "eu-west-1"
      scale_up_wait_duration: 1m
      scale_down_wait_duration: 5s

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 2m
        scale_down_duration: 10m

    - kind: limit
      config:
        max: 1000
        min: 2

  inputters:
  - name: sqs_queue_basic_tutorial_job_number
    description: "Get quantity based on the messages on the queue aplying a constant factor"
    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.eu-west-1.amazonaws.com/xxxxxxxxxxx/basic-tutorial-queue"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "eu-west-1"
    arrange:
      kind: constant_factor
      config:
        factor: 20
        round_type: "ceil"
```

Now we need to run our autoscaler with AWS credentials:

```bash
docker run  \
  --rm -it \
  -v `pwd`/ladder.yml:/etc/ladder/ladder.yml \
  -v `pwd`/basic_tutorial.yml:/etc/ladder/cfg-autoscalers/tutorial1/basic_tutorial.yml \
  -e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
  -e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
  -p 9094:9094 themotion/ladder
```

{{< note title="Note" >}}
You can use `-dry.run` flag to test it before autoscale nothing.
{{< /note >}}

It will wait the warmup before start running the autoscaler, you can access to its endpoints on port `9094`, for example: http://127.0.0.1:9094/check and it will give you information:

```json
{  
   "status":"Ok",
   "uptime":"1m45.514659061s",
   "check_ts":"2017-04-29 08:56:35.267076739 +0000 UTC",
   "version":"Ladder version 0.1.0, build master(695fa7b)",
   "healthy":{  
      "ladder_autoscaler":{  
         "basic_tutorial_autoscaler_sqs_asg":"running"
      }
   },
   "unhealthy":{  
      "ladder_autoscaler":{  

      }
   }
}
```

And that's it! now your platform is autoscaled. Easy right?