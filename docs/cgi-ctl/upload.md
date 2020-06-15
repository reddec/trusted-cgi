---
layout: default
title: upload
parent: Control util
nav_order: 201
---
## upload

From `0.3.1`

Upload the lambda content from the current directory to a remote instance of `trusted-cgi`.

Files defined in `.cgiignore` file will be ignored (uses `tar --exclude-form` syntax).

```
Usage:
  cgi-ctl [OPTIONS] upload [upload-OPTIONS]

Help Options:
  -h, --help          Show this help message

[upload command options]
      -l, --login=    Login name (default: admin) [$LOGIN]
      -p, --password= Password (default: admin) [$PASSWORD]
      -P, --ask-pass  Get password from stdin [$ASK_PASS]
      -u, --url=      Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
      -o, --uid=      Lambda UID [$UID]
          --input=    Directory (default: .) [$INPUT]
```


**Example 1** (lambda `e0ed902f-4a9c-4c29-870d-f343f330b6ab`)

```
cgi-ctl upload -o e0ed902f-4a9c-4c29-870d-f343f330b6ab --url https://example.com/ -P
```


will ask password for `admin` user, make archive and upload to the instance
