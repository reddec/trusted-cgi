---
layout: default
title: do
parent: Control util
nav_order: 204
---
## upload

From `0.3.2`

Invoke defined action(s) on the remote platform. If no actions provided for the utility, list of all
available actions will be printed.

```
Usage:
  cgi-ctl [OPTIONS] do [do-OPTIONS] [Actions...]

Help Options:
  -h, --help             Show this help message

[do command options]
      -l, --login=       Login name (default: admin) [$LOGIN]
      -p, --password=    Password (default: admin) [$PASSWORD]
      -P, --ask-pass     Get password from stdin [$ASK_PASS]
      -u, --url=         Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
          --ghost        Disable save credentials to user config dir [$GHOST]
          --independent  Disable read credentials from user config dir [$INDEPENDENT]
      -o, --uid=         Lambda UID (if empty - dirname of input will be used) [$UID]

[do command arguments]
  Actions:               action names
```


**Example** (from the remote instance, lambda `e0ed902f-4a9c-4c29-870d-f343f330b6ab`, from default python template):

```
cgi-ctl do --uid e0ed902f-4a9c-4c29-870d-f343f330b6ab --url https://example.com/ -P install
```

will ask password for `admin` user and then invoke `install` action for `e0ed902f-4a9c-4c29-870d-f343f330b6ab` lambda


For a [cloned](../clone) lambda it is enough to just call `do`:

```
cgi-ctl do install
```