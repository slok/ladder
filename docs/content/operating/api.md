---
date: 2017-01-01T11:08:00Z
title: API
menu:
  main:
    parent: Operating
    weight: 33
---


Ladder API at this moment only has one version `v1`, the default
entrypoint is `/api/v1` but this can be configured.

## Autoscalers

Autoscalers entrypoint are prefixed with `/autoscalers`, this entrypoints have
the actions that can be executed on autoscalers

### List autoscalers

This enpoint will return the present autoscalers and their state

* path: `/autoscalers`
* method: `GET`

#### Request

```bash
curl http://ladder.host/api/v1/autoscalers
```

#### Response:

Code: `200`
Body:

```json
{  
   "autoscalers":{  
      "asg1":{  
         "status":"running"
      },
      "asg2":{  
         "status":"running"
      },
      "asg3":{  
         "status":"stopped"
      },
      "asg4":{  
         "status":"running"
      }
   }
}

```

### Stop autoscaler for a period

Autoscalers state should be running, so the concept or stop for ever is not valid on Ladder
this enpoint will stop a running autoscaler for a period of time ([golang duration format](https://golang.org/pkg/time/#ParseDuration) valid

* path: `/autoscalers/{autoscaler_name}/stop/{duration}`
* method: `PUT`

For example stop `render_instances` autoscaler for 1:30h

#### Request

```bash
curl -XPUT http://ladder.host/api/v1/autoscalers/render_instances/stop/1h30m
```

#### Response when autoscaler running

Code: `202`
Body:

```json
{  
   "autoscaler":"render_instances",
   "msg":"Autoscaler stop request sent"
}
```
#### Response when autoscaler already stopped

Code: `409`
Body:

```json
{  
   "data":{  
      "autoscaler":"render_instances",
      "deadline":1485267342,
      "msg":"Autoscaler already stopped",
      "required-action":"Need to cancel current stop state first"
   },
   "error":"Autoscaler already stopped"
}
```

{{< note title="Note" >}}
Deadline is UTC unix epoch
{{< /note >}}

### Cancel an autoscaler stop action

this enpoint will cancel the stop state of an autoscaler

* path: `/autoscalers/{autoscaler_name}/cancel-stop`
* method: `PUT`

#### Request

```bash
curl -XPUT http://ladder.host/api/v1/autoscalers/render_instances/cancel-stop
```

#### Response when autoscaler stopped:

Code: `202`
Body:

```json
{  
   "autoscaler":"render_instances",
   "msg":"Autoscaler stop cancel request sent"
}
```

#### Response when autoscaler running:

Code: `400`
Body:

```json
{  
   "data":{  
      "autoscaler":"render_instances",
      "msg":"Autoscaler is not stopped"
   },
   "error":"Autoscaler is not stopped"
}
```
