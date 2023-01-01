---
layout: default
title: Manifest
parent: Usage
nav_order: 0
---
# Manifest

Manifest is the entrypoint for the server. File `manifest.json` is required for each
lambda.

Minimal manifest looks like 

```json
{
  "run": ["/bin/sh","./myscript.sh"]
}
```

where `/bin/sh` is runner and `./myscript.sh` is argument for it.


Example: for python with virtualenv with main script `app.py` it will look like:

```json
{
  "run": ["./venv/bin/python", "app.py"]
}
```
 
## Supported parameters

### Manifest

* **name** (optional, string): information field, a caption that will displayed in UI
* **description** (optional, string): information field, markdown based description, displayed in UI in overview
* **run** (required, array of string): command and arguments that will be executed (shell specific operations like pipes are not allowed)
* **output_headers** (optional, map of strings): output headers and values - key is header name, value is header value
* **input_headers** (optional, map of strings): input headers mapping, where key is header name and value is environment variable name to be fulfilled
* **query** (optional, map of strings): query (or form) mapping, where key is query parameter name and value is environment variable name to be fulfilled
* **environment** (optional, map of strings): environment variables that will be added to the lambda
* **method** (optional, string): allow requests only for specified HTTP method (POST, GET, etc..., but OPTIONS is not allowed)
* **method_env** (optional, string): map request path to specified environment variable
* **time_limit** (optional, time string): limit maximum execution time for the lambda. 
* **maximumPayload** (optional, number): limit incoming request size in bytes
* **cron** (option, array of `Cron`): scheduled actions
* **static** (optional, string): path to directory inside lambda to serve static files; if defined the GET and HEAD methods will not be available for handler

### Cron

* **cron** (required, string): cron tab expression (with seconds), [see scheduler doc](scheduler.md)
* **action** (required, string): target in Makefile to invoke, [see actions doc](actions.md)
* **time_limit**  (optional, time string): limit maximum execution time for the action



### Time string 

Uses [Go time.Duration](https://golang.org/pkg/time/#ParseDuration): string with suffixes:
 
* `ns` - nano seconds
* `us` - micro seconds
* `ms` - millisecond
* `s` - seconds
* `m` - minutes
* `h` - hours

Example: `1h30m25s`, `15s`

## Migration notice

### 0.3.3

* **aliases** (optional, array of string): aliases/links for the lambda, useful to make permanent URL, [see aliases doc](aliases.md)

Since `0.3.3` field `aliases` moved to platform level. Migration from `0.3.x` (x < 3) to `0.3.3` 
version should be done automatically after a restart.

### 0.3.5

* **allowed_ip** (optional, array of string): allow requests only from IP defined in the list
* **allowed_origin** (optional, array of string): allow requests only with `Origin` header value the list
* **tokens** (optional, array of string): allow requests only with `Authorization` header value the list; applicable only if `public: false`
* **public** (optional, boolean): if false, check all requests against `tokens`

Since `0.3.5` fields `allowed_ip`, `allowed_origin`, `tokens`, `public` moved to [policies](../administrating/policies.md). 
Migration from `0.3.x` (x < 5) to `0.3.5` version should be done automatically after a restart. 