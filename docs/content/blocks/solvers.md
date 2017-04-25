---
date: 2016-11-13T15:05:43Z
title: Solvers
menu:
  main:
    parent: Available blocks
    weight: 22
---

{{< note title="Note" >}}
You only will need a solver if you have multiple inputters, if not it will be
ignored although there is configured on the file
{{< /note >}}

## Dummy

Dummy solver will return the sum of all the inputs

### Name

`dummy`

### Options

No options

### Example

```yaml
solve:
  kind: dummy
```

{{< warning title="Warning" >}}Only used for testing{{< /warning >}}

## Bound

Bound solver will return the max or min of all the received inputters quantity.

### Name

`bound`

### Options

* `kind`: Can be `max` or `min`

### Example

```yaml
solve:
  kind: bound
  config:
    kind: max
```
