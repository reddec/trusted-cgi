---
layout: default
title: Development
nav_order: 999
---
# Development

The project has Linux-specific features and aimed to be run in Linux ecosystem.

Requirements for backend

* go 1.13 and upper - see [Go installation manual](https://golang.org/doc/install)
* make - should be available in your OS repository
* [git with LFS](https://git-lfs.github.com/) - for cloning binary UI

IDE:

* I am personally using professional [Jetbrains Goland](https://www.jetbrains.com/go/)
* Someone can prefer use free Visual Studio Code with Go plugin
* Or use Liteide

Quick project setup:

1. Install  [Go](https://golang.org/doc/install), setup `GOPATH` and add `GOPATH/bin` in your `PATH` variable (should be described in the link)
2. Clone project somewhere (not required only under `GOAPTH`, because the project uses modules): `git clone https://github.com/reddec/trusted-cgi.git`
2.1 Pull LFS files: `git lfs pull`
3. Change directory to the cloned folder (`cd trusted-cgi`) and build it: `go get -v ./cmd/...`
4. Test project build: `trusted-cgi --help` (version should be `dev`)
5. Run locally: `trusted-cgi --dev`

When you decide to change something and check - repeat 3, 4, 5, however in a matured IDE (like Goland) it's
possible to run `main.go` directly from UI that allows you to attach debugger if needed.

This is example of my runner in Goland (nothing complicated, but change working directory to your folder out of GIT tracking)

![image](https://user-images.githubusercontent.com/6597086/83396622-d9568b80-a42e-11ea-8be4-93f7b4cff0c2.png)
  

## Embedding UI

Requirements should be same as for UI [sub-repo](https://github.com/reddec/trusted-cgi-ui)

```shell
make clean
make embed_ui
```

## Regenerating API

In case of API changes, generated files should also be update by following command

```shell
make regen
```