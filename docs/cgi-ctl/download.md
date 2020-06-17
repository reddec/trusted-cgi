---
layout: default
title: download
parent: Control util
nav_order: 200
---
## download

From `0.3.1`

Download the lambda content from a remote instance of `trusted-cgi`.

```
Usage:
  cgi-ctl [OPTIONS] download [download-OPTIONS]

Help Options:
  -h, --help             Show this help message

[download command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -U, --uid=         Lambda UID [$UID]
      -o, --output=      Output data (- means stdout, empty means as UID) [$OUTPUT]
```

**Example 1** (from local dev instance, lambda `e0ed902f-4a9c-4c29-870d-f343f330b6ab`):

```
cgi-ctl download -i e0ed902f-4a9c-4c29-870d-f343f330b6ab
```

will create `e0ed902f-4a9c-4c29-870d-f343f330b6ab.tar.gz` archive.


**Example 2** (from the remote instance, same lambda):

```
cgi-ctl download -i e0ed902f-4a9c-4c29-870d-f343f330b6ab --url https://example.com/ -P
```

will ask password for `admin` user and then download archive

**Example 3** (from the remote instance, same lambda):

```
cgi-ctl download -i e0ed902f-4a9c-4c29-870d-f343f330b6ab --url https://example.com/ -P -o - | tar  zxf -
```

will ask password for `admin` user, download archive and unpack to current directory

