---
layout: default
title: Installation
parent: Administrating
nav_order: 1
---

{:toc:}

# Install

TL;DR;

* for production for debian servers - use `apt.reddec.net`
* locally or non-debian server - download binary and run
* for quick tests or for limited production - use docker image

## Debian/Ubuntu

Packages are served via Cloudflare R2 CDN (using [aptly](https://www.aptly.info/)) and signed by public key available
in https://apt.reddec.dev/archive.key.

> Note: please bear in mind it's non-commercial project and some rate-limits could be applied by me to CDN if the total
> bill will be too high.

<details>
<summary>Public key</summary>

```
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=w6TJ
-----END PGP PUBLIC KEY BLOCK-----
```

</details>

### APT

Add key

    wget -qO - https://apt.reddec.dev/archive.key | sudo tee /etc/apt/trusted.gpg.d/reddec-dev.asc 

Add repository

    sudo add-apt-repository 'deb https://apt.reddec.dev all main'

Install

    sudo apt install trusted-cgi

Available packages:

- `trusted-cgi` (meta package, contains both server and client)
- `trusted-cgi-server` (server only)
- `trusted-cgi-client` (client only)

## Deb files (manual)

Download the latest [release](https://github.com/reddec/trusted-cgi/releases).

Install your distribution:

* standard (basic templates supported): `sudo apt install ./trusted-cgi_0.3.7_linux_amd64.deb`
* minimal (actions will not work): `sudo apt install --no-install-recommends ./trusted-cgi_0.3.7_linux_amd64.deb`
* maximum (all pre-made templates available): `sudo apt install ./trusted-cgi_0.3.7_linux_amd64.deb php-cli nodejs npm`

Of course, you may install required packages later.

Inspect configuration file in `/etc/trusted-cgi/trusted-cgi.env`.

After any change in configuration file restart service: `sudo systemctl restart trusted-cgi`

By-default, the service will be available over http://127.0.0.1:3434 with credentials `admin/admin`

## Docker

**Notice:** due to docker nature it is impossible to make restrictions by IP.

* Pull image: `docker pull reddec/trusted-cgi`
* Run for test `docker run --rm -p 3434:3434 reddec/trusted-cgi`

There are several exposed parameters (see Dockerfile), however, data stored in `/data` and
initial admin password is `admin` (change it!).

The docker image contains pre-installed python3 (+requests), node js (+axios) and php to let experiment with default
functions.

There is light (around 8MB) docker image: `reddec/trusted-cgi:latest-light`. It contains only minimal set of
pre-installed
packages and could be useful to run pre-compiled binary functions or shell lambdas. Or to use as a base image.

## From source

Requirements:

* go 1.13+

Command: `go get -v -u github.com/reddec/trusted-cgi/cmd/...`

It will install both: `trusted-cgi` and control tool `cgi-ctl`.

See 'install from binary section' for the usage.

## Binary

Download suitable pre-compiled binary from [releases](https://github.com/reddec/trusted-cgi/releases)

Unpack archives to the PATH directory (ex: `/usr/local/bin`).

Use `trusted-cgi --help` to see help.
