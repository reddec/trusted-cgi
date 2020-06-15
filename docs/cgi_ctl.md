# CGI-CTL command help


`cgi-ctl` command aims to be a tool for helping developers interact with the platform without web UI.

`cgi-ctl` includes into distribution starting from version 0.3.1 and could be obtained independently
via source code or pre-built binary ([see installation](installation.md))

## init bare

From `0.3.0`

Initialize a basic lambda in a current directory.

```
Usage:
  cgi-ctl [OPTIONS] init bare [bare-OPTIONS]

Help Options:
  -h, --help             Show this help message

[bare command options]
          --git          Enable Git [$GIT]
      -d, --description= Description (default: Bare project) [$DESCRIPTION]
      -P, --private      Mark as private [$PRIVATE]
      -t, --time-limit=  Time limit for execution (default: 10s) [$TIME_LIMIT]
      -p, --max-payload= Maximum payload (default: 8192) [$MAX_PAYLOAD]

```

## download

From `0.3.1`

Download the lambda content from a remote instance of `trusted-cgi`.

```
Usage:
  cgi-ctl [OPTIONS] download [download-OPTIONS]

Help Options:
  -h, --help          Show this help message

[download command options]
      -l, --login=    Login name (default: admin) [$LOGIN]
      -p, --password= Password (default: admin) [$PASSWORD]
      -P, --ask-pass  Get password from stdin [$ASK_PASS]
      -u, --url=      Trusted-CGI endpoint (default: http://127.0.0.1:3434/) [$URL]
      -i, --uid=      Lambda UID [$UID]
      -o, --output=   Output data (- means stdout, empty means as UID) [$OUTPUT]
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
