---
layout: default
title: PoliciesAPI
parent: API
---

# PoliciesAPI

API for managing policies


* [PoliciesAPI.List](#policiesapilist) - List all policies
* [PoliciesAPI.Create](#policiesapicreate) - Create new policy
* [PoliciesAPI.Remove](#policiesapiremove) - Remove policy
* [PoliciesAPI.Update](#policiesapiupdate) - Update policy definition
* [PoliciesAPI.Apply](#policiesapiapply) - Apply policy for the resource
* [PoliciesAPI.Clear](#policiesapiclear) - Clear applied policy for the lambda



## PoliciesAPI.List

List all policies

* Method: `PoliciesAPI.List`
* Returns: `[]application.Policy`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "PoliciesAPI.List",
    "params" : []
}
EOF
```

### Policy


| Json | Type | Comment |
|------|------|---------|
| id | `string` |  |
| definition | `PolicyDefinition` |  |
| lambdas | `types.JsonStringSet` |  |

### Token


Signed JWT

## PoliciesAPI.Create

Create new policy

* Method: `PoliciesAPI.Create`
* Returns: `*application.Policy`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | policy | `string` |
| 2 | definition | `PolicyDefinition` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "PoliciesAPI.Create",
    "params" : []
}
EOF
```

### Policy


| Json | Type | Comment |
|------|------|---------|
| id | `string` |  |
| definition | `PolicyDefinition` |  |
| lambdas | `types.JsonStringSet` |  |

### PolicyDefinition


| Json | Type | Comment |
|------|------|---------|
| allowed_ip | `types.JsonStringSet` |  |
| allowed_origin | `types.JsonStringSet` |  |
| public | `bool` |  |
| tokens | `map[string]string` |  |

### Token


Signed JWT

## PoliciesAPI.Remove

Remove policy

* Method: `PoliciesAPI.Remove`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | policy | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "PoliciesAPI.Remove",
    "params" : []
}
EOF
```

### Token


Signed JWT

## PoliciesAPI.Update

Update policy definition

* Method: `PoliciesAPI.Update`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | policy | `string` |
| 2 | definition | `PolicyDefinition` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "PoliciesAPI.Update",
    "params" : []
}
EOF
```

### PolicyDefinition


| Json | Type | Comment |
|------|------|---------|
| allowed_ip | `types.JsonStringSet` |  |
| allowed_origin | `types.JsonStringSet` |  |
| public | `bool` |  |
| tokens | `map[string]string` |  |

### Token


Signed JWT

## PoliciesAPI.Apply

Apply policy for the resource

* Method: `PoliciesAPI.Apply`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | lambda | `string` |
| 2 | policy | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "PoliciesAPI.Apply",
    "params" : []
}
EOF
```

### Token


Signed JWT

## PoliciesAPI.Clear

Clear applied policy for the lambda

* Method: `PoliciesAPI.Clear`
* Returns: `bool`

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
    "method" : "PoliciesAPI.Clear",
    "params" : []
}
EOF
```

### Token


Signed JWT