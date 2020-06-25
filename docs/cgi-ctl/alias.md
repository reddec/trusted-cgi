---
layout: default
title: alias
parent: Control util
nav_order: 206
---

# alias

Creates, removes or print [aliases](../../usage/aliases) for a lambda.

* no flags - print all known aliases for the lambda
* with `-d` flag - delete defined aliases from the lambda
* without `-d` flag - add aliases to the lambda

```
Usage:
  cgi-ctl [OPTIONS] alias [alias-OPTIONS] [Aliases...]

Help Options:
  -h, --help             Show this help message

[alias command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -U, --uid=         Lambda UID [$UID]
      -d, --delete       delete links, otherwise add [$DELETE]
          --keep         do not update (if it exists) local manifest file [$KEEP]

[alias command arguments]
  Aliases:               links/aliases names
```

**Example** - local instance after [clone](../clone), create aliases

```
cgi-ctl alias alias1 alias2
```

**Example** - local instance after [clone](../clone), print aliases

```
cgi-ctl alias
```


**Example** - local instance after [clone](../clone), remove one alias

```
cgi-ctl alias -d alias1
```