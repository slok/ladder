---
date: 2016-11-13T18:02:13Z
title: Configuration
menu:
  main:
    parent: Operating
    weight: 31
---



## Configuration schema

Ladder configuration is splitted in 2 main blocks, one the global configuration and
a multiple autoscalers blocks. The main entrypoint configuration of Ladder will point
to the other configuration files where the autoscalers are configured.

For example our main configuration is `ladder.yml` (by default will be this):

```yaml
global:
  metrics_path: /metrics
  api_v1_path: /api/v1
  interval: 30s
  warmup: 3m
  scaling_wait_timeout: 3m

autoscaler_files:
  - cfg-autoscalers/services/amis/*.yml
  - cfg-autoscalers/services/main_cluster/*.yml
  - cfg-autoscalers/clusters/*.yml
```

This file points to our autoscalers that will reside there, see the file structure:

```bash
./ladder.yml
./cfg-autoscalers/
├── clusters
│   ├── main.yml
│   └── wrong.json
└── services
    ├── amis
    │   └── render.yml
    └── main_cluster
        ├── infra.yml
        └── video.yml
```

As you see the paths are path matchers, that `wrong.json` will be ignored.

## Global configuration

The global configuration is the configuration that will be applied to Ladder as
a program or by default to all the autoscalers depending on the setting

It starts with `global:`

* `metrics_path`: The path where the metrics can be retrieved, by default `/metrics`
* `config_path`: The path where the loaded configuration files can be retrieved, by default: `/config`
* `health_check_path`: The path where the health check will be listening, by default `/check`
* `api_v1_path`: The prefix path where the API v1 enpoints will be listening, by default `/api/v1`
* `interval`: The interval the autoscaler will run the iteration process, by default `30s`
* `warmup`: The time the autoscaler will wait for the first scalation execution
(gathering, solving... will occur), by default `30s`
* `scaling_wait_timeout`: The time that will wait before giving timeout when a
correct scalation starts the process of waiting until the target has scaled `2m`

## Autoscalers cofiguration files

Autoscaler configuration files have one or multiple autoscalers per file, thats up to you
and how do you organize the autoscalers. For example this could be a very simple autoscaling file:

```yaml
autoscalers:
- name: autoscaler1
  description: "As1"

  scale:
    kind: aws_autoscaling_group
    config:
      aws_region: "us-west-2"
      auto_scaling_group_name: "slok-ECSAutoScalingGroup-1PNI4RX8BD5XU"

  inputters:
  - name: aws_sqs_constant_factor
    description: "Will get a number based on the queue messages and a constant factor division"
    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.us-west-2.amazonaws.com/016386521566/slok-render-jobs"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "us-west-2"

    arrange:
      kind: constant_factor
      config:
        factor: 10
        round_type: "ceil"
# Autoscaler 2
- name: autoscaler2
  description: "As2"

  scale:
    kind: aws_autoscaling_group
    config:
      aws_region: "us-west-2"
      auto_scaling_group_name: "slok-ECSAutoScalingGroup-1PNI4RX8BD5XU"

  inputters:
  - name: aws_sqs_constant_factor
    description: "Will get a number based on the queue messages and a constant factor division"
    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.us-west-2.amazonaws.com/016386521566/slok-render-jobs"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "us-west-2"

    arrange:
      kind: constant_factor
      config:
        factor: 10
        round_type: "ceil"
```

## Configuration example

This is a real example of multiple autoscalers (services, clusters...), the file structure is this:

```bash
.
├── cfg-autoscalers
│   ├── factory
│   │   ├── analytics.yml
│   │   ├── processing.yml
│   │   ├── rendering.yml
│   │   └── transcoding.yml
│   └── infra
│       └── cluster.yml
└── ladder.yml
```


### `ladder.yml`

```yaml
global:
  metrics_path: /metrics
  interval: 30s
  warmup: 3m
  scaling_wait_timeout: 2m

autoscaler_files:
  - "cfg-autoscalers/factory/*.yml"
  - "cfg-autoscalers/infra/*.yml"
```

### `cfg-autoscalers/infra/cluster.yml`

```yaml
autoscalers:
- name: ECS_cluster
  disabled: false
  description: >
    ECS autoscaler will set the correct number of instances on an ECS autoscaling
    group based on the memory or cpu reserved percentage of that autoscaler cloudwatch
    metrics
  interval: 1m
  scaling_wait_timeout: 6m
  scale:
    kind: aws_autoscaling_group
    config:
      auto_scaling_group_name: "prod-ECSAutoScalingGroup-F4VEQM9FVL2U"
      aws_region: "eu-west-1"
      scale_up_wait_duration: 3m
      scale_down_wait_duration: 1m30s

  solve:
    kind: bound
    config:
      kind: max

  filters:
    - kind: ecs_running_tasks
      config:
        aws_region: eu-west-1
        cluster_name: prod-ECSCluster1-OIA8GT0KCY6X
        max_pending_tasks_allowed: 0
        max_checks: 10
        error_on_max_checks: true
        when: scale_down

    - kind: scaling_kind_interval
      config:
        scale_up_duration: 3m
        scale_down_duration: 6m

    - kind: limit
      config:
        max: 300
        min: 8

  inputters:
  - name: memory_reserved_based_input
    description: >
      This input will arrange the required number of instances on a cluster based
      on an ECS cluster memory reservation
    gather:
      kind: prometheus_metric
      config:
        addresses:
          - http://prometheus.prod.bi.themotion.lan
          - http://prometheus2.prod.bi.themotion.lan
        query: cluster:container_memory_remaining_for_reservation:bytes{type="ecs"}
    arrange:
      kind: threshold
      config:
        scaleup_threshold: 16106127360
        scaledown_threshold: 48318382080
        scaleup_percent: 40
        scaledown_percent: 20
        scaleup_max_quantity: 10
        scaledown_max_quantity: 2
        scaleup_min_quantity: 1
        scaledown_min_quantity: 1
        inverse: true

  - name: cpu_reserved_based_input
    description: >
      This input will arrange the required number of instances on a cluster based
      on an ECS cluster cpu reservation
    gather:
      kind: prometheus_metric
      config:
        addresses:
          - http://prometheus.prod.bi.themotion.lan
          - http://prometheus2.prod.bi.themotion.lan
        query: cluster:container_cpu_remaining_for_reservation:cpu_shares{type="ecs"}
    arrange:
      kind: threshold
      config:
        scaleup_threshold: 8192
        scaledown_threshold: 24576
        scaleup_percent: 40
        scaledown_percent: 20
        scaleup_max_quantity: 10
        scaledown_max_quantity: 2
        scaleup_min_quantity: 1
        scaledown_min_quantity: 1
        inverse: true

```

### `cfg-autoscalers/factory/rendering.yml`

```yaml
autoscalers:
- name: render_instances
  disabled: false
  description: "Render autoscaler will autoscale machines based on the SQS video rendering jobs"
  interval: 1m
  scaling_wait_timeout: 6m

  scale:
    kind: aws_autoscaling_group
    config:
      auto_scaling_group_name: "prod-ami-render-AMIAutoScalingGroup-1X8U7Q03UC4BC"
      aws_region: "eu-west-1"
      scale_up_wait_duration: 1m
      scale_down_wait_duration: 5s

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 30s
        scale_down_duration: 20m

    - kind: limit
      config:
        max: 2500
        min: 3

  inputters:
  - name: render_instances_based_on_jobs_queues
    description: "Get quantity based on the jobs length with a constant factor"

    gather:
      kind: prometheus_metric
      config:
        addresses:
          - http://prometheus.prod.bi.themotion.lan
          - http://prometheus2.prod.bi.themotion.lan
        query: number_pending_jobs{queue="render"}
    arrange:
      kind: constant_factor
      config:
        factor: 5
        round_type: "ceil"
```

### `cfg-autoscalers/factory/processing.yml`

```yaml
autoscalers:
- name: processor_service
  disabled: false
  description: "Procesor service will autoscale instances based on the SQS processing jobs"
  interval: 30s
  scaling_wait_timeout: 5m

  scale:
    kind: aws_ecs_service
    config:
      aws_region: eu-west-1
      cluster_name: prod-ECSCluster1-OIA8GT0KCY6X
      service_name: processor

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 1m
        scale_down_duration: 5m

    - kind: limit
      config:
        max: 1000
        min: 2

  inputters:
  - name: service_instances_based_on_jobs_queue
    description: "Get quantity based on the jobs length with a constant factor"

    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.eu-west-1.amazonaws.com/843176375373/prod-processing"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "eu-west-1"
    arrange:
      kind: constant_factor
      config:
        factor: 10
        round_type: "ceil"
```

### `cfg-autoscalers/factory/transcoding.yml`

```yaml
autoscalers:
- name: transcoder_service
  disabled: false
  description: "Transcodder service will autoscale instances based on the SQS video rendering jobs"
  interval: 30s
  scaling_wait_timeout: 5m

  scale:
    kind: aws_ecs_service
    config:
      aws_region: eu-west-1
      cluster_name: prod-ECSCluster1-OIA8GT0KCY6X
      service_name: transcoder

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 1m
        scale_down_duration: 5m

    - kind: limit
      config:
        max: 800
        min: 2

  inputters:
  - name: service_instances_based_on_jobs_queue
    description: "Get quantity based on the jobs length with a constant factor"

    gather:
      kind: prometheus_metric
      config:
        addresses:
          - http://prometheus.prod.bi.themotion.lan
          - http://prometheus2.prod.bi.themotion.lan
        query: number_pending_jobs{queue="transcode"}
    arrange:
      kind: constant_factor
      config:
        factor: 10
        round_type: "ceil"

```

### `cfg-autoscalers/factory/analytics.yml`

```yaml
autoscalers:
- name: analytics_mine_service
  disabled: false
  description: "Analytics Mine autoscales based on the number of import jobs present in the queue"
  interval: 30s
  scaling_wait_timeout: 5m

  scale:
    kind: aws_ecs_service
    config:
      aws_region: eu-west-1
      cluster_name: prod-ECSCluster1-OIA8GT0KCY6X
      service_name: analytics-mine

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 1m
        scale_down_duration: 5m

    - kind: limit
      config:
        max: 50
        min: 0

  inputters:
  - name: service_instances_based_on_jobs_queue
    description: "Get quantity based on the jobs length with a constant factor"

    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.eu-west-1.amazonaws.com/843176375373/prod-analytics-batches"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "eu-west-1"
    arrange:
      kind: constant_factor
      config:
        factor: 5
        round_type: "ceil"

- name: analytics_forge_service
  disabled: false
  description: "Analytics Forge autoscales based on the number of jobs in the queue"
  interval: 30s
  scaling_wait_timeout: 5m

  scale:
    kind: aws_ecs_service
    config:
      aws_region: eu-west-1
      cluster_name: prod-ECSCluster1-OIA8GT0KCY6X
      service_name: analytics-forge

  filters:
    - kind: scaling_kind_interval
      config:
        scale_up_duration: 1m
        scale_down_duration: 5m

    - kind: limit
      config:
        max: 5
        min: 1

  inputters:
  - name: service_instances_based_on_jobs_queue
    description: "Get quantity based on the jobs length with a constant factor"

    gather:
      kind: aws_sqs
      config:
        queue_url: "https://sqs.eu-west-1.amazonaws.com/843176375373/prod-analytics-jobs"
        queue_property: "ApproximateNumberOfMessages"
        aws_region: "eu-west-1"
    arrange:
      kind: constant_factor
      config:
        factor: 360
        round_type: "ceil"

```
