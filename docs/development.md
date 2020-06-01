# Development

The project has Linux-specific features and aimed to be run in Linux ecosystem.

Requirements for backend

* go 1.13 and upper - see [Go installation manual](https://golang.org/doc/install)
* make - should be available in your OS repository
* git - for cloning

IDE:

* I am personally using professional [Jetbrains Goland](https://www.jetbrains.com/go/)
* Someone can prefer use free Visual Studio Code with Go plugin
* Or use Liteide

## Embedding UI

Requirements should be same as for UI [sub-repo](https://github.com/reddec/trusted-cgi-ui)

```shell
make clean
make embed_ui
```