---
layout: default
title: update manifest
parent: Control util
nav_order: 207
---

# update manifest

Pulls remote lambda manifest and saves it locally. Useful to update only the configuration, while
keeping files remotely unsynchronized with local environment.

```
Usage:
  cgi-ctl [OPTIONS] update manifest [manifest-OPTIONS]

Help Options:
  -h, --help             Show this help message

[manifest command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -U, --uid=         Lambda UID (if empty - dirname of input will be used) [$UID]
```
