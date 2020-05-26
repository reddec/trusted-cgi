# API




* [API.Login](#apilogin) - Login user by username and password. Returns signed JWT
* [API.ChangePassword](#apichangepassword) - Change password for the user
* [API.Create](#apicreate) - Create new app (lambda)
* [API.Config](#apiconfig) - Project configuration
* [API.Apply](#apiapply) - Apply new configuration and save it
* [API.AllTemplates](#apialltemplates) - Get all templates without filtering
* [API.CreateFromTemplate](#apicreatefromtemplate) - Create new app/lambda/function using pre-defined template
* [API.Upload](#apiupload) - Upload content from .tar.gz archive to app and call Install handler (if defined)
* [API.Download](#apidownload) - Download content as .tar.gz archive from app
* [API.Push](#apipush) - Push single file to app
* [API.Pull](#apipull) - Pull single file from app
* [API.List](#apilist) - List available apps (lambdas) in a project
* [API.Remove](#apiremove) - Remove app and call Uninstall handler (if defined)
* [API.Templates](#apitemplates) - Templates with filter by availability including embedded
* [API.Files](#apifiles) - Files in func dir
* [API.Info](#apiinfo) - Info about application
* [API.Update](#apiupdate) - Update application manifest
* [API.CreateFile](#apicreatefile) - Create file or directory inside app
* [API.RemoveFile](#apiremovefile) - Remove file or directory
* [API.RenameFile](#apirenamefile) - Rename file or directory
* [API.GlobalStats](#apiglobalstats) - Global last records
* [API.Stats](#apistats) - Stats for the app
* [API.Actions](#apiactions) - Actions available for the app
* [API.Invoke](#apiinvoke) - Invoke action in the app (if make installed)



## API.Login

Login user by username and password. Returns signed JWT

* Method: `API.Login`
* Returns: `*Token`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | login | `string` |
| 1 | password | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Login",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.ChangePassword

Change password for the user

* Method: `API.ChangePassword`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | password | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.ChangePassword",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Create

Create new app (lambda)

* Method: `API.Create`
* Returns: `*application.App`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Create",
    "params" : []
}
EOF
```
### App

| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
### Token

```go
type Token struct {
}
```

## API.Config

Project configuration

* Method: `API.Config`
* Returns: `*application.ProjectConfig`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Config",
    "params" : []
}
EOF
```
### ProjectConfig

| Json | Type | Comment |
|------|------|---------|
| user | `string` |  |
| untar | `[]string` |  |
| tar | `[]string` |  |
### Token

```go
type Token struct {
}
```

## API.Apply

Apply new configuration and save it

* Method: `API.Apply`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | config | `ProjectConfig` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Apply",
    "params" : []
}
EOF
```
### ProjectConfig

| Json | Type | Comment |
|------|------|---------|
| user | `string` |  |
| untar | `[]string` |  |
| tar | `[]string` |  |
### Token

```go
type Token struct {
}
```

## API.AllTemplates

Get all templates without filtering

* Method: `API.AllTemplates`
* Returns: `[]*TemplateStatus`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.AllTemplates",
    "params" : []
}
EOF
```
### TemplateStatus

| Json | Type | Comment |
|------|------|---------|
| available | `bool` |  |
### Token

```go
type Token struct {
}
```

## API.CreateFromTemplate

Create new app/lambda/function using pre-defined template

* Method: `API.CreateFromTemplate`
* Returns: `*application.App`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | templateName | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.CreateFromTemplate",
    "params" : []
}
EOF
```
### App

| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
### Token

```go
type Token struct {
}
```

## API.Upload

Upload content from .tar.gz archive to app and call Install handler (if defined)

* Method: `API.Upload`
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
    "method" : "API.Upload",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Download

Download content as .tar.gz archive from app

* Method: `API.Download`
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
    "method" : "API.Download",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Push

Push single file to app

* Method: `API.Push`
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
    "method" : "API.Push",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Pull

Pull single file from app

* Method: `API.Pull`
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
    "method" : "API.Pull",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.List

List available apps (lambdas) in a project

* Method: `API.List`
* Returns: `[]*application.App`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.List",
    "params" : []
}
EOF
```
### App

| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
### Token

```go
type Token struct {
}
```

## API.Remove

Remove app and call Uninstall handler (if defined)

* Method: `API.Remove`
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
    "method" : "API.Remove",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Templates

Templates with filter by availability including embedded

* Method: `API.Templates`
* Returns: `[]*Template`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Templates",
    "params" : []
}
EOF
```
### Template

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| description | `string` |  |
### Token

```go
type Token struct {
}
```

## API.Files

Files in func dir

* Method: `API.Files`
* Returns: `[]*File`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | name | `string` |
| 2 | dir | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.Files",
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

## API.Info

Info about application

* Method: `API.Info`
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
    "method" : "API.Info",
    "params" : []
}
EOF
```
### App

| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
| manifest | `types.Manifest` |  |
### Token

```go
type Token struct {
}
```

## API.Update

Update application manifest

* Method: `API.Update`
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
    "method" : "API.Update",
    "params" : []
}
EOF
```
### App

| Json | Type | Comment |
|------|------|---------|
| uid | `string` |  |
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
| post_clone | `string` |  |
### Token

```go
type Token struct {
}
```

## API.CreateFile

Create file or directory inside app

* Method: `API.CreateFile`
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
    "method" : "API.CreateFile",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.RemoveFile

Remove file or directory

* Method: `API.RemoveFile`
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
    "method" : "API.RemoveFile",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.RenameFile

Rename file or directory

* Method: `API.RenameFile`
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
    "method" : "API.RenameFile",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.GlobalStats

Global last records

* Method: `API.GlobalStats`
* Returns: `[]stats.Record`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | limit | `int` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "API.GlobalStats",
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

## API.Stats

Stats for the app

* Method: `API.Stats`
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
    "method" : "API.Stats",
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

## API.Actions

Actions available for the app

* Method: `API.Actions`
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
    "method" : "API.Actions",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## API.Invoke

Invoke action in the app (if make installed)

* Method: `API.Invoke`
* Returns: `bool`

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
    "method" : "API.Invoke",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```