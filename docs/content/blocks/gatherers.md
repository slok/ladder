---
date: 2016-11-13T15:05:29Z
title: Gatherers
menu:
  main:
    parent: Available blocks
    weight: 20
---

## Dummy

Dummy gatherer always will return a constant quantity

### Name

`dummy`

### Options

* `quantity`: The quantity to return always

### Example

```yaml
gather:
  kind: dummy
  config:
    quantity: 0
```

{{< warning title="Warning" >}}Only used for testing{{< /warning >}}

## Random

Random gatherer will return a random number between a max and min bounds

### Name

`random`

### Options

* `max_limit`: The max limit of the random (not included)
* `min_limit`: The min limit of the random (included)

### Example

```yaml
gather:
  kind: random
  config:
    max_limit: 10
    min_limit: 0
```

## SQS property

SQS gatherer will return the quantity of the number of messages of a queue, this
number can be one of the available SQS prperties

### Name

`aws_sqs`

### Options

* `queue_url`: The SQS queue URL
* `queue_property`: The property to get the message number, can be one of these 3:
    * `ApproximateNumberOfMessages`
    * `ApproximateNumberOfMessagesNotVisible`
    * `ApproximateNumberOfMessagesDelayed`
* `aws_region`: The region of AWS where the SQS queue lives

### Example

```yaml
gather:
  kind: aws_sqs
  config:
    queue_url: "https://sqs.us-west-2.amazonaws.com/016386521566/slok-render-jobs"
    queue_property: "ApproximateNumberOfMessages"
    aws_region: "us-west-2"
```

## Cloudwatch metric

Cloudwatchmetric gatherer will return a current metric of a given query for the
last minute aggregation, this let us know curren values (approx.) about AWS metrics.

### Name

`aws_cloudwatch_metric`

### Options

* `aws_region`: The region of AWS where the cloudwatch metrics
* `metric_name`: The name of the metric(`CPUReservation`, `CPUCreditBalance`, `DiskWriteBytes`...)
* `namespace`: The namespace of the metric (`AWS/ELB`, `AWS/EC2`...)
* `statistic`: The statistic type of the metric. Check them at:     https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Statistic
* `unit`: The unit type of the metric, Check them at:   https://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#Unit
* `offset`: 0 or negative time duration to apply to the metrics query, for example `-30s` will get the metrics from -1'30'' to -30'' metrics from now
* `dimensions`: dimensions are like prometheus labels, is a list of dicts having `name` and `value` with this the metric wil be filtered

### Example

```yaml
gather:
  kind: aws_cloudwatch_metric
  config:
    aws_region: "us-west-2"
    metric_name: "CPUReservation"
    namespace: "AWS/ECS"
    statistic: "Maximum"
    unit: "Percent"
    offset: "-30s"
    dimensions:
    - name: "ClusterName"
      value: "slok-ECSCluster1-15OBYPKBNXIO6"
```

## Prometheus metric

Prometheus metric gatherer is one of the most powerful gatherers, not because of the gatherer itself, but for
the amazing Prometheus query API. This gatherer should work with single sample vectors, other type of results
from Prometheus will error, for example a vector with length greater than 1 or a Matrix result.

Prometheus gatherer accepts different prometheus so it can fallback to a different prometheus to get the metric.
[HA Prometheus](https://prometheus.io/docs/introduction/faq/#can-prometheus-be-made-highly-available?) infrastructure is usually made by 2 equal prometheis that are independent one each other

{{< note title="Note" >}}
You should query precalculated metrics (recording rules), this will speed up the query and will be more
readable
{{< /note >}}

### Name

`prometheus_metric`
### Options

* `addresses`: The addresses of the prometheus endpoint
* `query`: The query that will be send to prometheus

### Example

```yaml
gather:
  kind: prometheus_metric
  config:
    addresses:
      - http://prometheus.prod.bi.themotion.lan
      - http://prometheus2.prod.bi.themotion.lan
      - http://prometheus3.prod.bi.themotion.lan
    query: max(service:container_memory_usage:percent{service="prometheus"})
```
