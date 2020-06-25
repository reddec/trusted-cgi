---
layout: default
title: LambdaAPI
parent: API
---

# LambdaAPI

API for lambdas


* [LambdaAPI.Upload](#lambdaapiupload) - Upload content from .tar.gz archive to app and call Install handler (if defined)
* [LambdaAPI.Download](#lambdaapidownload) - Download content as .tar.gz archive from app
* [LambdaAPI.Push](#lambdaapipush) - Push single file to app
* [LambdaAPI.Pull](#lambdaapipull) - Pull single file from app
* [LambdaAPI.Remove](#lambdaapiremove) - Remove app and call Uninstall handler (if defined)
* [LambdaAPI.Files](#lambdaapifiles) - Files in func dir
* [LambdaAPI.Info](#lambdaapiinfo) - Info about application
* [LambdaAPI.Update](#lambdaapiupdate) - Update application manifest
* [LambdaAPI.CreateFile](#lambdaapicreatefile) - Create file or directory inside app
* [LambdaAPI.RemoveFile](#lambdaapiremovefile) - Remove file or directory
* [LambdaAPI.RenameFile](#lambdaapirenamefile) - Rename file or directory
* [LambdaAPI.Stats](#lambdaapistats) - Stats for the app
* [LambdaAPI.Actions](#lambdaapiactions) - Actions available for the app
* [LambdaAPI.Invoke](#lambdaapiinvoke) - Invoke action in the app (if make installed)
* [LambdaAPI.Link](#lambdaapilink) - Make link/alias for app
* [LambdaAPI.Unlink](#lambdaapiunlink) - Remove link



## LambdaAPI.Upload

Upload content from .tar.gz archive to app and call Install handler (if defined)

* Method: `LambdaAPI.Upload`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | tarGz | `[]byte` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Upload",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Download

Download content as .tar.gz archive from app

* Method: `LambdaAPI.Download`
* Returns: `[]byte`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Download",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Push

Push single file to app

* Method: `LambdaAPI.Push`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | file | `string` |
| 3 | content | `[]byte` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Push",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Pull

Pull single file from app

* Method: `LambdaAPI.Pull`
* Returns: `[]byte`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | file | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Pull",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Remove

Remove app and call Uninstall handler (if defined)

* Method: `LambdaAPI.Remove`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Remove",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Files

Files in func dir

* Method: `LambdaAPI.Files`
* Returns: `[]types.File`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | dir | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Files",
    "params" : []
}
EOF
```

### File


| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| is_dir | `bool` |  |

### Token


Signed JWT

## LambdaAPI.Info

Info about application

* Method: `LambdaAPI.Info`
* Returns: `*application.Definition`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Info",
    "params" : []
}
EOF
```

### Definition


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| aliases | `types.JsonStringSet` |  |
| manifest | `types.Manifest` |  |

### Token


Signed JWT

## LambdaAPI.Update

Update application manifest

* Method: `LambdaAPI.Update`
* Returns: `*application.Definition`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | manifest | `Manifest` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Update",
    "params" : []
}
EOF
```

### Definition


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| aliases | `types.JsonStringSet` |  |
| manifest | `types.Manifest` |  |

### Manifest


| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| description | `string` |  |
| run | `[]string` |  |
| output_headers | `map[string]string` |  |
| input_headers | `map[string]string` |  |
| query | `map[string]string` |  |
| environment | `map[string]string` |  |
| method | `string` |  |
| method_env | `string` |  |
| path_env | `string` |  |
| time_limit | `JsonDuration` |  |
| maximum_payload | `int64` |  |
| allowed_ip | `JsonStringSet` |  |
| allowed_origin | `JsonStringSet` |  |
| public | `bool` |  |
| tokens | `map[string]string` |  |
| cron | `[]Schedule` |  |
| static | `string` |  |

### Token


Signed JWT

## LambdaAPI.CreateFile

Create file or directory inside app

* Method: `LambdaAPI.CreateFile`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | path | `string` |
| 3 | dir | `bool` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.CreateFile",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.RemoveFile

Remove file or directory

* Method: `LambdaAPI.RemoveFile`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | path | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.RemoveFile",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.RenameFile

Rename file or directory

* Method: `LambdaAPI.RenameFile`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | oldPath | `string` |
| 3 | newPath | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.RenameFile",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Stats

Stats for the app

* Method: `LambdaAPI.Stats`
* Returns: `[]stats.Record`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | limit | `int` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Stats",
    "params" : []
}
EOF
```

### Record


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| error | `string` |  |
| request | `types.Request` |  |
| begin | `time.Time` |  |
| end | `time.Time` |  |

### Token


Signed JWT

## LambdaAPI.Actions

Actions available for the app

* Method: `LambdaAPI.Actions`
* Returns: `[]string`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Actions",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Invoke

Invoke action in the app (if make installed)

* Method: `LambdaAPI.Invoke`
* Returns: `string`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | action | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Invoke",
    "params" : []
}
EOF
```

### Token


Signed JWT

## LambdaAPI.Link

Make link/alias for app

* Method: `LambdaAPI.Link`
* Returns: `*application.Definition`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | uid | `string` |
| 2 | alias | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Link",
    "params" : []
}
EOF
```

### Definition


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| aliases | `types.JsonStringSet` |  |
| manifest | `types.Manifest` |  |

### Token


Signed JWT

## LambdaAPI.Unlink

Remove link

* Method: `LambdaAPI.Unlink`
* Returns: `*application.Definition`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | alias | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "LambdaAPI.Unlink",
    "params" : []
}
EOF
```

### Definition


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| aliases | `types.JsonStringSet` |  |
| manifest | `types.Manifest` |  |

### Token


Signed JWT