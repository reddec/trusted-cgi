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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

## LambdaAPI.Files

Files in func dir

* Method: `LambdaAPI.Files`
* Returns: `[]*File`

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
| is_dir | `bool` |  |
| name | `string` |  |

### Token


```go
type Token struct {
}
```

## LambdaAPI.Info

Info about application

* Method: `LambdaAPI.Info`
* Returns: `*application.App`

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

### App


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
| git | `bool` |  |

### Token


```go
type Token struct {
}
```

## LambdaAPI.Update

Update application manifest

* Method: `LambdaAPI.Update`
* Returns: `*application.App`

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

### App


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
| git | `bool` |  |

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
| aliases | `JsonStringSet` |  |
| cron | `[]Schedule` |  |

### Token


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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
| input | `[]byte` |  |
| output | `[]byte` |  |
| error | `string` |  |
| code | `int` |  |
| method | `string` |  |
| remote | `string` |  |
| origin | `string` |  |
| uri | `string` |  |
| token | `string` |  |
| begin | `time.Time` |  |
| end | `time.Time` |  |

### Token


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

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


```go
type Token struct {
}
```

## LambdaAPI.Link

Make link/alias for app

* Method: `LambdaAPI.Link`
* Returns: `*application.App`

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

### App


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
| git | `bool` |  |

### Token


```go
type Token struct {
}
```

## LambdaAPI.Unlink

Remove link

* Method: `LambdaAPI.Unlink`
* Returns: `*application.App`

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

### App


| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
| git | `bool` |  |

### Token


```go
type Token struct {
}
```