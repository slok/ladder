---
date: 2016-11-13T18:02:22Z
title: Metrics
menu:
  main:
    parent: Operating
    weight: 32
---

## Prometheus

Ladder serves prometheus metrics on `/metrics` by default, you can override this
on the global configuration like this:

```yaml
metrics_path: /my/awesome/metrics
```

For now this are the available metrics:

* `ladder_gatherer_quantity`
* `ladder_gatherer_duration_histogram_ms`
* `ladder_gatherer_errors_total`
* `ladder_inputter_quantity`
* `ladder_inputter_duration_histogram_ms`
* `ladder_inputter_errors_total`
* `ladder_solver_quantity`
* `ladder_solver_duration_histogram_ms`
* `ladder_solver_errors_total`
* `ladder_scaler_current_quantity`
* `ladder_scaler_current_duration_histogram_ms`
* `ladder_scaler_current_errors_total`
* `ladder_scaler_quantity`
* `ladder_scaler_duration_histogram_ms`
* `ladder_scaler_errors_total`
* `ladder_autoscaler_iterations_total`
* `ladder_autoscaler_errors_total`
* `ladder_autoscaler_duration_histogram_ms`
* `ladder_autoscaler_running`

## Grafana dashboard

With the metrics that Ladder exposes to Prometheus, using along with Grafana
you can have a very nice dashboard where you can see the state of Ladder.

Here you can download the [dashboard](/data/ladder-dashboard.json)

![Grafana dashboard](/img/grafana.png)
