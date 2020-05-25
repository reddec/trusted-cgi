# Trusted-CGI

[![license](https://img.shields.io/github/license/reddec/trusted-cgi.svg)](https://github.com/reddec/trusted-cgi)
[![](https://godoc.org/github.com/reddec/trusted-cgi?status.svg)](http://godoc.org/github.com/reddec/trusted-cgi)
[![donate](https://img.shields.io/badge/help_by️-donate❤-ff69b4)](http://reddec.net/about/#donate)
[![Download](https://api.bintray.com/packages/reddec/debian/trusted-cgi/images/download.svg)](https://bintray.com/reddec/debian/trusted-cgi/_latestVersion)

![](https://bintray-binary-objects-or-production.s3-accelerate.amazonaws.com/80ee75735ebc642670140a263e7e94f32fb8ce932933626ef3c4812006295af0)

Lightweight self-hosted lambda/applications/cgi/serverless-functions engine. 

**Why?**
 
Because I want to write small handlers that will be 99% of time just do nothing. I am already paying for the cheapest
Digital Ocean (thanks guys for your existence) and do not want to pay additionally to Lambda providers like Google/Amazon/Azure.

I also tried self hosted solutions based on k3s but it too heavy for 1GB server (yep, it is, don't believe in marketing).

So, 'cause I am a Developer I decided to make my own wheels ;-)

**Idea behind**

Idea came from past: CGI. In a beginning of Internet, people have being making a simple scripts that receives incoming bytes over STDIN 
(standard input) and writes to STDOUT (standard output). The application server (aka CGI server), accepts clients,
invokes scripts and redirects socket input/output to the script. There are a lot of details here but this is brief explanation.

After more than 20 years the world spin around and arrived to the beginning: serverless functions/lambda and so on.
It is almost CGI, except scripts became a docker containers, and we need much more servers to do the same things as before.

So let's cut the corners a bit: we have a trusted developer (our self, company workers - means it's not arbitrary clients), 
so we don't need a heavy restriction for the application, so let's throw away docker and another heavy staff.

Add some piece of **security**: inbound IP, inbound origins, tokens....

Add monitoring of hits and history details....

Add neat Web UI file browsing with edit functions...

Add playground where you can test your scripts....

Add nice logo, license everything under MIT and you will get Trusted-CGI.  
 

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
echo "deb https://dl.bintray.com/reddec/debian all main" | sudo tee -a /etc/apt/sources.list
sudo apt update
sudo apt install trusted-cgi
```

Configuration files will be placed under `/etc/trusted-cgi`, functions files under `/var/trusted-cgi`,
systemd service will be launched as `trusted-cgi` and all new services will be run under `trusted-cgi` system
user.

## Docker

* Pull image: `docker pull reddec/trusted-cgi`
* Run for test `docker run --rm -p 3434:3434 reddec/trusted-cgi`

There are several exposed parameters (see Dockerfile), however, data stored in `/data` and
initial admin password is `admin` (change it!).

The docker image contains pre-installed python3 (+requests), node js (+axios) and php to let experiment with default
functions.

# Development

## Embedding UI

```shell script
make clean
make embed_ui
`

## TODO

* Upload/download tarball
* CLI control