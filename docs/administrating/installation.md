---
layout: default
title: Installation
parent: Administrating
nav_order: 1
---
{:toc:}
# Install

TL;DR;

* for production for debian servers - use bintray repository (recommend)
* locally or non-debian server - download binary and run
* for quick tests or for limited production - use docker image

## Debian/Ubuntu

Add the repository (needed only once)

```bash
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61
echo "deb https://dl.bintray.com/reddec/debian all main" | sudo tee /etc/apt/sources.list.d/trusted-cgi.list
sudo apt update
```

Install your distribution:

* standard (basic templates supported): `sudo apt install trusted-cgi`
*  minimal (actions will not work): `sudo apt install --no-install-recommends trusted-cgi`
* maximum (all pre-made templates available): `sudo apt install trusted-cgi php-cli nodejs npm`

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
