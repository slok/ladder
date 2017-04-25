---
date: 2016-11-13T15:05:35Z
title: Arrangers
menu:
  main:
    parent: Available blocks
    weight: 21
---

## Dummy

Dummy gatherer always will return a constant quantity

### Name

`dummy`

### Options

* `quantity`: The quantity to return always

### Example

```yaml
arrange:
  kind: dummy
  config:
    quantity: 0
```

{{< warning title="Warning" >}}Only used for testing{{< /warning >}}

## In list

In list arranger will arrange a new quantity to scale up/down based on two lists
When the quantity received is in `match_upscale` list it will scale the current
quantity by the `match_up_magnitude` percent. The same for the downscale.

For example scaling up 100 by a magnitude of 50 it will result in 150. Downscale
by 20 will result in 80

### Name

`in_list`

### Options

* `match_downscale`: The downscale list of quantities (numbers only allowed)
* `match_upscale`: The upscale list of quantities (numbers only allowed)
* `match_up_magnitude`: The upscale magnitude (scalation of the current by %)
* `match_down_magnitude`:  The downscale magnitude (scalation of the current by %)

### Example

```yaml
arrange:
  kind: in_list
  config:
    match_downscale: [0, 2, 4, 6, 8]
    match_upscale: [1, 2, 5, 7, 9]
    match_up_magnitude:   200
    match_down_magnitude: 50
```

## Constant factor

Constant factor will arrange a new quantity based on a constant factor division.
For example an input of 500 with a constant factor of 10 will result on a 50
quantity. When a floating point result is arranged it will round
the result up/down based on the `round_type` option

###  Name

`constant_factor`

### Options:
* `factor`: The factor for the division of the input
* `round_type`: The type of rounding, available ones:
    * `ceil`: Rounds up
    * `floor`: Rounds down

### Example

```yaml
arrange:
  kind: constant_factor
  config:
    factor: 10
    round_type: "ceil"
```

## Threshold

Threshold will arrange a new quantity based on upper ond lower thresholds, it
will scale up & down based on the percent of the current cuantity until the input
gets between the 2 thresholds, it can scale up and down in a different way.

{{< note title="Note" >}}
By default this arranger will scale up when the value received is above `scaleup_threshold`, and scale down when the value
is below `scaledown_threshold`, to invert this and scale up when the value is below `scaleup_threshold` and scale down when the
value is greater than `scaledown_threshold`, you need to use the `inverse` setting
{{< /note >}}

### Name

`threshold`

### Options
* `scaleup_threshold`: The threshold to start scaling up
* `scaledown_threshold`: The threshold to start scaling up
* `scaleup_percent`: The percent of current value that will be add when scaling up is triggered
* `scaledown_percent`: The percent of current value that will be substract when scaling down is triggered
* `scaleup_max_quantity`: The max quantity of the scaling up value (the delta get from `scaleup_percent` with current value)
* `scaledown_max_quantity`: The max quantity of the scaling down value (the delta get from `scaledown_percent` with current value)
* `scaleup_min_quantity`: The min quantity of the scaling up value (the delta get from `scaleup_percent` with current value)
* `scaledown_min_quantity`: The min quantity of the scaling down value (the delta get from `scaledown_percent` with current value)
* `inverse`: By default this arranger will scale up when the value received is above the `scaleup_threshold` threshold and will scale down when the value is below `scaledown_threshold`, if inverse is true it will invert this and will scaleup when the value is below
the threshold and scale down when is above the threshold

### Example

```yaml
arrange:
  kind: threshold
  config:
    scaleup_threshold: 80
    scaledown_threshold: 70
    scaleup_percent: 10
    scaledown_percent: 10
    scaleup_max_quantity: 30
    scaledown_max_quantity: 10
    scaleup_min_quantity: 2
    scaledown_min_quantity: 1
    inverse: false
```

{{< note title="Note" >}}
To upscale or downscale with a fixed absolute value instead of percent you can set to 0 (or omit) scaling percents and
the min and max of the scaling mode to a fixed value
{{< /note >}}

### Example with fixed value

Downscale with a fixed value of 2 and upscale with a fixed value of 5

```yaml
arrange:
  kind: threshold
  config:
    scaleup_threshold: 80
    scaledown_threshold: 70
    #scaleup_percent: 0
    #scaledown_percent: 0
    scaleup_max_quantity: 5
    scaleup_min_quantity: 5
    scaledown_max_quantity: 2
    scaledown_min_quantity: 2
    innverse: false
```

{{< note title="Note" >}}
With this arranger (and the [`scaling_kind_interval`]({{< relref "blocks/filters.md#scaling-kind-interval" >}}) filterer) we met the requirements for a dynamic and independent growth/reduction
like the described on Netflix tech post: [`Auto scaling in amazon cloud`](http://techblog.netflix.com/2012/01/auto-scaling-in-amazon-cloud.html)
{{< /note >}}
