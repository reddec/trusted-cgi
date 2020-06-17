---
layout: default
title: Control util
nav_order: 600
has_children: true
---

`cgi-ctl` command aims to be a tool for helping developers interact with the platform without web UI.

`cgi-ctl` includes into distribution starting from version 0.3.1 and could be obtained independently
via source code or pre-built binary ([see installation](../administrating/installation))

Always use `cgi-ctl --help` for actual help.

## Stored configuration

Since `0.3.2`, enabled by-default, disable by `--independed`, `--ghost`

### Credentials

cgi-ctl utility tries to keep login information in a user configuration dir under trusted-cgi-ctl subdir
 (linux: `~/.config/trusted-cgi-ctl`).
 
Each file in the directory represents configuration for each host, where hostname is a filename (where `:` replaced to `_`).

For example, for local instance `127.0.0.1:3434` will be generated file `127.0.0.1_3434` with following content

```json
{
  "login": "admin",
  "password": "YWRtaW4="
}
```

where `password` is base64 encoded password as was entered by user

* To disable **load** the configuration file use `--independed` flag
* To disable **save** the configuration file use `--ghost` flag

### Remote URL

In file `.cgictl.json` will be saved remote configuration: URL.

Example, after a `clone` operation from local dev instance:

```json
{
  "url": "http://127.0.0.1:3434/"
}
``` 

## General login sequence

1. Go to (2) if flag `--independed` set
   * read remote URL from `.cgictl.json` file if possible
   * read config from `~/.config/trusted-cgi-ctl/<host>` if possible
   * on success - disable `--ask-pass` flag
2. If `--ask-pass` set - ask for a password from STDIN without echo.
3. Login and token
4. If flag `--ghost` not set, save credentials to `~/.config/trusted-cgi-ctl/<host>`

## General UID search

1. Use `-U, --uid` flag if presented;
2. Otherwise, read `.cgictl.json` file is exists and use `uid` field (if not empty);
3. Otherwise, Use current directory name as UI