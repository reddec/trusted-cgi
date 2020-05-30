# Trusted-CGI

[![license](https://img.shields.io/github/license/reddec/trusted-cgi.svg)](https://github.com/reddec/trusted-cgi)
[![](https://godoc.org/github.com/reddec/trusted-cgi?status.svg)](http://godoc.org/github.com/reddec/trusted-cgi/application)
[![donate](https://img.shields.io/badge/help_by️-donate❤-ff69b4)](http://reddec.net/about/#donate)
[![Download](https://api.bintray.com/packages/reddec/debian/trusted-cgi/images/download.svg)](https://bintray.com/reddec/debian/trusted-cgi/_latestVersion)

![](https://bintray-binary-objects-or-production.s3-accelerate.amazonaws.com/80ee75735ebc642670140a263e7e94f32fb8ce932933626ef3c4812006295af0)

Lightweight self-hosted lambda/applications/cgi/serverless-functions engine. 

[see docs](https://trusted-cgi.reddec.net)

Features:

* No specific requirements: just one binary. Working "as-is"
* One-click new lambda with public link and handler. Available immediately.
* Rich API
* Security: user switch, IP restrictions, Origin restrictions, tokens ....
* Time limits
* Permanent links (aliases)
* Actions - independent instruction that could be run via UI/API on server
* Scheduler: run actions in cron-tab like style
* ... etc - [see docs](https://trusted-cgi.reddec.net) 
 

# Installation

## Play locally

Just download and run `trusted-cgi --dev`

## Direct to server (recommended)

Recommended: ubuntu LTS x64 server

0. Add bintray key `sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61`
1. Download from [releases](https://github.com/reddec/trusted-cgi/releases) page, or (better) use bintray repo
[![Download](https://api.bintray.com/packages/reddec/debian/trusted-cgi/images/download.svg)](https://bintray.com/reddec/debian/trusted-cgi/_latestVersion)
2. `apt update` - update repos (optional since 18.04 and you used bintray repo)
3. `apt install trusted-cgi` or for minimal `apt install --no-install-recommends trusted-cgi`  

For Ubuntu (should be for all LTS)

```bash
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 379CE192D401AB61
echo "deb https://dl.bintray.com/reddec/debian all main" | sudo tee /etc/apt/sources.list.d/trusted-cgi.list
sudo apt update
sudo apt install trusted-cgi
```

Configuration files will be placed under `/etc/trusted-cgi`, functions files under `/var/trusted-cgi`,
systemd service will be launched as `trusted-cgi` and all new services will be run under `trusted-cgi` system
user.

## Docker

Notice: due to docker nature it is impossible to make restrictions by IP.

* Pull image: `docker pull reddec/trusted-cgi`
* Run for test `docker run --rm -p 3434:3434 reddec/trusted-cgi`

There are several exposed parameters (see Dockerfile), however, data stored in `/data` and
initial admin password is `admin` (change it!).

The docker image contains pre-installed python3 (+requests), node js (+axios) and php to let experiment with default
functions.

# Docs and features

* [Manifest](docs/manifest.md) - main and mandatory entrypoint for the lambda
* [Actions](docs/actions.md) - arbitrary actions that could be invoked by UI or by scheduler
* [Scheduler](docs/scheduler.md) - cron-like scheduling system to automatically call actions by time
* [Aliases](docs/aliases.md) - permanent links and aliases/links
* [Security](docs/security.md) - security and restrictions
* [GIT repo](docs/git_repo.md) - using GIT repo as a function

# Actions

If function contains Makefile and installed make, it is possible to invoke targets over UI/API (called Actions). Useful
for installing dependencies or building.

# URL

Each function contains at least one URL: `<base URL>/a/<UID>` and any number of unique aliases/links `<base URL>/l/<LINK NAME>`.

Links are useful to make a public links and dynamically migrate between real implementations (functions). For ex:
you made a GitHub hook processor in Python language, than changed your mind and switched to PHP function. Instead of 
updating link in GitHub repo (that could be a hassle if you spread it everywhere) you can change just a link.

Important! Security settings and restrictions will be used from new functions.

# Templates

## Embedded

### Python 3

Host requirements:

* make
* python3
* python3-venv

### Node

Host requirements:

* make
* node
* npm

### PHP

Host requirements:

* php

### Nim lang

Host requirements:

* make
* nim
* nimble

# Development

## Embedding UI

```shell script
make clean
make embed_ui
`

## TODO

* Upload/download tarball
* CLI control