# ProjectAPI

Remove link


* [ProjectAPI.Config](#projectapiconfig) - Get global configuration
* [ProjectAPI.SetUser](#projectapisetuser) - Change effective user
* [ProjectAPI.AllTemplates](#projectapialltemplates) - Get all templates without filtering
* [ProjectAPI.List](#projectapilist) - List available apps (lambdas) in a project
* [ProjectAPI.Templates](#projectapitemplates) - Templates with filter by availability including embedded
* [ProjectAPI.Stats](#projectapistats) - Global last records
* [ProjectAPI.Create](#projectapicreate) - Create new app (lambda)
* [ProjectAPI.CreateFromTemplate](#projectapicreatefromtemplate) - Create new app/lambda/function using pre-defined template



## ProjectAPI.Config

Get global configuration

* Method: `ProjectAPI.Config`
* Returns: `*Settings`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "ProjectAPI.Config",
    "params" : []
}
EOF
```
### Settings

| Json | Type | Comment |
|------|------|---------|
| user | `string` |  |
| public_key | `string` |  |
### Token

```go
type Token struct {
}
```

## ProjectAPI.SetUser

Change effective user

* Method: `ProjectAPI.SetUser`
* Returns: `*Settings`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | user | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "ProjectAPI.SetUser",
    "params" : []
}
EOF
```
### Settings

| Json | Type | Comment |
|------|------|---------|
| user | `string` |  |
| public_key | `string` |  |
### Token

```go
type Token struct {
}
```

## ProjectAPI.AllTemplates

Get all templates without filtering

* Method: `ProjectAPI.AllTemplates`
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
    "method" : "ProjectAPI.AllTemplates",
    "params" : []
}
EOF
```
### TemplateStatus

| Json | Type | Comment |
|------|------|---------|
| name | `string` |  |
| description | `string` |  |
| available | `bool` |  |
### Token

```go
type Token struct {
}
```

## ProjectAPI.List

List available apps (lambdas) in a project

* Method: `ProjectAPI.List`
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
    "method" : "ProjectAPI.List",
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

## ProjectAPI.Templates

Templates with filter by availability including embedded

* Method: `ProjectAPI.Templates`
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
    "method" : "ProjectAPI.Templates",
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

## ProjectAPI.Stats

Global last records

* Method: `ProjectAPI.Stats`
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
    "method" : "ProjectAPI.Stats",
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

## ProjectAPI.Create

Create new app (lambda)

* Method: `ProjectAPI.Create`
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
    "method" : "ProjectAPI.Create",
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

## ProjectAPI.CreateFromTemplate

Create new app/lambda/function using pre-defined template

* Method: `ProjectAPI.CreateFromTemplate`
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
    "method" : "ProjectAPI.CreateFromTemplate",
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