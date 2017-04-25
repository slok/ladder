---
date: 2016-11-13T15:05:51Z
title: Scalers
menu:
  main:
    parent: Available blocks
    weight: 24
---

## Dummy
Dummy scaler will return changed or not if the current quantity is same/different
of the new quantity, if different it will be the new current

### Name

`dummy`
### Options:

* `wait_duration`: The time to wait when waiting after scaling, by default `0ms`

### Example

```yaml
scale:
  kind: dummy
  config:
    wait_duration: 5s
```

{{< warning title="Warning" >}}Only used for testing{{< /warning >}}

## Stdout

Dummy scaler will return changed or not if the current quantity is same/different
of the new quantity, if different it will be the new current and will put a message
with the action made (upscale or downscale)

### Name

`stdout`

### Options

* `message_prefix`: The prefix for the messages

### Example

```yaml
scale:
  kind: stdout
  config:
    message_prefix: "[SCALER]"
```

## Auto scaling group

Auto scaling group scaler will set the desired instances of the group to a the
new input, if the new quantity is the same as the current it will do nothing.


### Name

`aws_autoscaling_group`

### Options

* `aws_region`: The AWS region where the auto scaling group lives
* `auto_scaling_group_name`: The name of the autoscalingr group
* `scale_up_wait_duration`: The time to wait after scaling up (to give time to the machines to stabilize)
* `scale_down_wait_duration`: The time to wait after scaling down
* `force_min_max`: boolean that will set the min and max instances properties on the asg if true
* `remaining_closest_hour_limit_duration`:If an instance has been running for N minutes in the last hour and this time is higher than
the minutes reminaing in this setting for the current running hour then the instance can be downscaled. Example:
`remaining_closest_hour_limit_duration` is 10m, we have 10 instances, now want 5, 2 of them have been running in the last running hour for
50m or more and 8 for less than 50m, the scaler will set 8 instances and not 5, because only 2 meet the requirements of >=50m
running.
* `max_no_downscale_rch_limit`: The maximum times the filter refering to `remaining_closest_hour_limit_duration` didn't downscale any
number, after this maximum times the filter will not activate and will downscale as it didn't apply any filtering

{{< note title="Note" >}}
If `scale_up_wait_duration` or `scale_down_wait_duration` are not defined, the scaler will wait until the desired
number of scaled instnaces met the running instances on the ASG
{{< /note >}}

{{< note title="Note" >}}
If `remaining_closest_hour_limit_duration` is set to `0[smh]` or is missing, it will be disabled
{{< /note >}}

{{< note title="Note" >}}
If `remaining_closest_hour_limit_duration` logic doesn't downscale any instance for `max_no_downscale_rch_limit` opt iterations, it will downscale to the desired ones
no matter running time of the instances, this way if there are problems with the timing calculation, pace check or whatever, it will continue downscaling as a regular iteration
{{< /note >}}

{{< note title="Note" >}}
AWS invoices for a full hour, `remaining_closest_hour_limit_duration` will enable a way of maximizing the use of scaled instances,
only downscaling to the number of instances required by the limit enables the [downscale policies](http://docs.aws.amazon.com/autoscaling/latest/userguide/as-instance-termination.html#default-termination-policy) of AWS.
{{< /note >}}


### Requirements

{{< note title="Note" >}}
It will need `autoscaling:UpdateAutoScalingGroup` AWS permission policy, for example:
{{< /note >}}

```json
{  
   "Version":"2012-10-17",
   "Statement":[  
      {  
         "Action":"autoscaling:UpdateAutoScalingGroup",
         "Resource":"*",
         "Effect":"Allow"
      }
   ]
}
```

### Example

```yaml
scale:
  kind: aws_autoscaling_group
  config:
    aws_region: "us-west-2"
    auto_scaling_group_name: "slok-ECSAutoScalingGroup-1PNI4RX8BD5XU"  
    scale_up_wait_duration: 5m
    scale_down_wait_duration: 15s
    force_min_max: true
    remaining_closest_hour_limit_duration: 10m
    max_no_downscale_rch_limit: 180

```

## ECS service

ECS service scaler will set the number of desired instances of an ECS service.
if the new quantity is the same as the current it will do nothing,

### Name

`aws_ecs_service`

### Options

* `aws_region`: The AWS region where the auto scaling group lives
* `cluster_name`: The name of the ECS cluster
* `service_name`: The name of the service in the ECS cluster

### Requirements

{{< note title="Note" >}}
It will need `ecs:UpdateService` AWS permission policy, for example:
{{< /note >}}

```json
{  
   "Version":"2012-10-17",
   "Statement":[  
      {  
         "Action":"ecs:UpdateService",
         "Resource":"*",
         "Effect":"Allow"
      }
   ]
}
```


### Example

```yaml
scale:
  kind: aws_ecs_service
  config:
    aws_region: us-west-2
    cluster_name: slok-ECSCluster1-15OBYPKBNXIO6
    service_name: alertmanager
```
