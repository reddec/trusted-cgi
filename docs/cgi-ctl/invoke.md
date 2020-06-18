---
layout: default
title: invoke
parent: Control util
nav_order: 209
---

## invoke

Invoke remote lambda for test. 

```
Usage:
  cgi-ctl [OPTIONS] invoke [invoke-OPTIONS]

Help Options:
  -h, --help              Show this help message

[invoke command options]
      -U, --uid=          Lambda UID [$UID]
      -i, --input=        input file that will be used as body (- or empty is stdin) (default: -) [$INPUT]
      -o, --output=       output file for response (- or empty is stdout) (default: -) [$OUTPUT]
      -g, --get           use GET method instead of POST (body will be ignored) [$GET]
      -t, --token=        add authorization token [$TOKEN]
      -O, --origin=       add origin header [$ORIGIN]
      -C, --content-type= set content-type header (default: application/json) [$CONTENT_TYPE]
      -H, --header=       custom headers [$HEADER]
      -f, --field=        set JSON field (input will be ignored) [$FIELD]
      -v, --verbose       show logs [$VERBOSE]
```

Better use in a [cloned](../clone) or [created](../create) lambda.

**Example** basic call

```
echo '{"name": "reddec"}' | cgi-ctl invoke
```

**Example** call with params

```
cgi-ctl invoke -f name:reddec
```
