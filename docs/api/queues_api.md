---
layout: default
title: QueuesAPI
parent: API
---

# QueuesAPI

API for managing queues


* [QueuesAPI.Create](#queuesapicreate) - Create queue and link it to lambda and start worker
* [QueuesAPI.Remove](#queuesapiremove) - Remove queue and stop worker
* [QueuesAPI.Linked](#queuesapilinked) - Linked queues for lambda
* [QueuesAPI.List](#queuesapilist) - List of all queues
* [QueuesAPI.Assign](#queuesapiassign) - Assign lambda to queue (re-link)



## QueuesAPI.Create

Create queue and link it to lambda and start worker

* Method: `QueuesAPI.Create`
* Returns: `*application.Queue`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | queue | `Queue` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "QueuesAPI.Create",
    "params" : []
}
EOF
```

### Queue


| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| target | `string` |  |
| retry | `int` |  |
| interval | `types.JsonDuration` |  |

### Token


Signed JWT

## QueuesAPI.Remove

Remove queue and stop worker

* Method: `QueuesAPI.Remove`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | name | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "QueuesAPI.Remove",
    "params" : []
}
EOF
```

### Token


Signed JWT

## QueuesAPI.Linked

Linked queues for lambda

* Method: `QueuesAPI.Linked`
* Returns: `[]application.Queue`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | lambda | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "QueuesAPI.Linked",
    "params" : []
}
EOF
```

### Queue


| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| target | `string` |  |
| retry | `int` |  |
| interval | `types.JsonDuration` |  |

### Token


Signed JWT

## QueuesAPI.List

List of all queues

* Method: `QueuesAPI.List`
* Returns: `[]application.Queue`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "QueuesAPI.List",
    "params" : []
}
EOF
```

### Queue


| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| target | `string` |  |
| retry | `int` |  |
| interval | `types.JsonDuration` |  |

### Token


Signed JWT

## QueuesAPI.Assign

Assign lambda to queue (re-link)

* Method: `QueuesAPI.Assign`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | name | `string` |
| 2 | lambda | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "QueuesAPI.Assign",
    "params" : []
}
EOF
```

### Token


Signed JWT