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
* Queues and retries
* ... etc - [see docs](https://trusted-cgi.reddec.net) 


P.S

There is minimal version of trusted-cgi: [nano-run](https://github.com/reddec/nano-run). Check it out - it DevOps friendly with configuration-first approach (ie easier to use for infrastructure-as-a-code).

# Installation

Since `0.3.3` Linux, Darwin and even Windows OS supported: pre-built binaries could be found in [releases](https://github.com/reddec/trusted-cgi/releases)

TL;DR;

* for production for debian servers - use bintray repository (recommend)
* locally or non-debian server - [download binary](https://github.com/reddec/trusted-cgi/releases) and run
* for quick tests or for limited production - use docker image (`docker run --rm -p 3434:3434 reddec/trusted-cgi`)

See [installation manual](https://trusted-cgi.reddec.net/administrating/installation/)

# Overview 

The process flow is quite straightforward: one light daemon in background listens for requests and launches scripts/apps
on demand. An executable shall read standard input (stdin) for request data and write a response to standard output (stdout).

Technically any script/application that can parse STDIN and write something to STDOUT should be capable of the execution.

Trusted-cgi designed keeping in mind that input and output data is quite small and contains structured data (json/xml),
however, there are no restrictions on the platform itself.

Key differences with classic CGI:

* Only request body is being piped to scripts input (CGI pipes everything, and application has to parse it by itself - it could be very not trivial and slow (it depends))
* Request headers, form fields, and query params are pre-parsed by the platform and can be passed as an environment variable (see mapping)
* Response headers are pre-defined in manifest

Due to changes, it's possible to make the simplest script with JSON input and output like this:

```python
import sys
import json

request = json.load(sys.stdin) # read and parse request
response = ['hello', 'world']  # do some logic and make response
json.dump(response, sys.stdout)  # send it to client
```  

Keep in mind, the platform also adds a growing number of new features - see features.

**target audience**

It's best (but not limited) for

* for hobby projects
* for experiments
* for projects with a low number of requests: webhooks, scheduled processing, etc..
* for a project working on low-end machines: raspberry pi, cheapest VPS, etc..

However, if your projects have overgrown the platform limitations, it should be quite easy to migrate to any other solutions, because
most low-level details are hidden and could be replaced in a few days (basically - just wrap script to HTTP service)  

Also, it is possible to scale the platform performance by just launching the same instances of the platform
with a shared file system (or docker images) with a balancer in front of it.


# Contributing

The platform is quite simple Golang project with Vue + Quasar frontend 
and should be easy for newcomers. Caveats and tips for backend check [here](https://trusted-cgi.reddec.net/development)

For UI check [sub-repo](https://github.com/reddec/trusted-cgi-ui)

Any PR (docs, code, styles, features, ...) will be very helpful!
