---
layout: default
title: clone
parent: Control util
nav_order: 203
---
## upload

From `0.3.2`

Download, unpack and setup local copy of lambda.

Automatically creates file `.cgictl.json` and adds it to `.cgiignore` (if not presented).

```
Usage:
  cgi-ctl [OPTIONS] clone [clone-OPTIONS]

Help Options:
  -h, --help             Show this help message

[clone command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -U, --uid=         Lambda UID [$UID]
      -o, --output=      Output directory (empty - same as UID) [$OUTPUT]
```


**Example** (from the remote instance, lambda `e0ed902f-4a9c-4c29-870d-f343f330b6ab`):

```
cgi-ctl clone -i e0ed902f-4a9c-4c29-870d-f343f330b6ab --url https://example.com/ -P
```

will ask password for `admin` user and then download and unpack an archive to `e0ed902f-4a9c-4c29-870d-f343f330b6ab` dir

When you decide to upload changes, change dir to cloned lambda and invoke

```
cgi-ctl upload
```