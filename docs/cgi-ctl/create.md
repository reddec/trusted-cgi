---
layout: default
title: create
parent: Control util
nav_order: 205
---
## create

Creates a new lambda on the remote platform. Initializes local environment: 
.cgiignore, [manifest.json](../../usage/manifest) and .cgictl.json files.

Uses default server template (usually - bare minimal).

From `0.3.3`

```
Usage:
  cgi-ctl [OPTIONS] create [create-OPTIONS] [Name]

Help Options:
  -h, --help             Show this help message

[create command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]

[create command arguments]
  Name:                  project directory
```

**Example 1** - create using local dev instance

```
cgi-ctl create example-1
```

will create lambda and initialize local directory `example-1`

**Example 2** - create using remote instance

```
cgi-ctl create --url https://example.com -P example-2
```
