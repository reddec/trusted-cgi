---
layout: default
title: Actions
parent: Usage
nav_order: 1
---
# Actions

Actions is optional arbitrary commands defined in [Makefile](https://www.gnu.org/software/make/manual/make.html#Rule-Example) as targets and can be invoked
by UI, admin API or during template cloning operations.

Main purpose is to prepare environment or function out of general flow procedure (HTTP call): 
build binary, download dependencies, etc.

UI:
 
1. click to any created application
2. click to actions tab


Basic example - update from git.

For example, your lambda source code hosted somewhere in Git repo and you already
initialized function from Git clone.

You can put `update` action in Makefile that will pull latest `master` branch.

`makefile`
```makefile
update:
	git pull origin master
```

That's it!


You will in UI/{Lambda}/Actions button `update`. If you will push it, the `update`
target will be invoked.

You also can schedule automatic update in a `Schedule` tab!

Bonus: if you used `create from git` button for new lambda in UI, the `update` target
will be automatically generated for your convenience.