# Trusted-CGI

[![license](https://img.shields.io/github/license/reddec/trusted-cgi.svg)](https://github.com/reddec/trusted-cgi)
[![](https://godoc.org/github.com/reddec/trusted-cgi?status.svg)](http://godoc.org/github.com/reddec/trusted-cgi/application)
[![donate](https://img.shields.io/badge/help_by️-donate❤-ff69b4)](http://reddec.net/about/#donate)
[![Download](https://api.bintray.com/packages/reddec/debian/trusted-cgi/images/download.svg)](https://bintray.com/reddec/debian/trusted-cgi/_latestVersion)

Lightweight self-hosted lambda/applications/cgi/serverless-functions engine. 

![Download](./assets/interaction.svg)

<iframe width="560" height="315" src="https://www.youtube.com/embed/GjqhQXlOdWQ" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

**Idea behind**

The idea came from the past: CGI. At the beginning of the Internet, people have been making a simple script that receives incoming bytes over STDIN 
(standard input) and writes to STDOUT (standard output). The application server (aka CGI server), accepts clients,
invokes scripts and redirects socket input/output to the script. There are a lot of details here but this is a brief explanation.

After more than 20 years the world spin around and arrived at the beginning: serverless functions/lambda and so on.
It is almost CGI, except scripts became docker containers, and we need many more servers to do the same things as before.

So let's cut the corners a bit: we have a trusted developer (our self, company workers - means it's not arbitrary clients), 
so we don't need a heavy restriction for the application, so let's throw away docker and another heavy staff.

## Docs and features

* [Manifest](manifest.md) - main and mandatory entrypoint for the lambda
* [Actions](actions.md) - arbitrary actions that could be invoked by UI or by scheduler
* [Scheduler](scheduler.md) - cron-like scheduling system to automatically call actions by time
* [Aliases](aliases.md) - permanent links and aliases/links
* [Security](security.md) - security and restrictions
* [GIT repo](git_repo.md) - using GIT repo as a function

**High-level components diagram**

![Download](./assets/trusted-cgi-overview.svg)

## Why I did it?
 
Because I want to write small handlers that will be 99% of the time just do nothing. I am already paying for the cheapest
Digital Ocean (thanks guys for your existence) and do not want to pay additionally to Lambda providers like Google/Amazon/Azure.

I also tried self-hosted solutions based on k3s but it too heavy for 1GB server (yep, it is, don't believe in marketing).

So, 'cause I am a developer I decided to make my own wheels ;-)

# Installation

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