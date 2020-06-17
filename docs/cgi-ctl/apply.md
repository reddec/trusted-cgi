---
layout: default
title: apply
parent: Control util
nav_order: 208
---

## apply

Pushes local manifest to the remote platform and applies settings without a restart.

```
Usage:
  cgi-ctl [OPTIONS] apply [apply-OPTIONS]

Help Options:
  -h, --help             Show this help message

[apply command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -U, --uid=         Lambda UID [$UID]
```
