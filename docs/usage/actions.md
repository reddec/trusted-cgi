---
layout: default
title: Actions
parent: Usage
nav_order: 1
---
# Actions

Actions are optional arbitrary commands defined in a [Makefile](https://www.gnu.org/software/make/manual/make.html#Rule-Example) as targets and can be invoked
by UI, admin API or during template cloning operations.

The main purpose is to prepare the environment or a function out of general flow procedure (HTTP call): 
build binary, download dependencies, etc.

UI:
 
1. click on any created application
2. click on the `Actions` tab


Basic example - update from git.

For example, your lambda source code is hosted somewhere in a Git repo and you already
initialized the function from Git clone.

You can put an `update` action in the Makefile that will pull the latest `master` branch.

`makefile`
```makefile
update:
	git pull origin master
```

That's it!


You will see the button `update` in the `Actions` tab. If you push it, the `update`
target will be invoked.

You also can schedule an automatic update in the `Schedule` tab!

Bonus: if you used the `create from git` button for a new lambda in the UI, the `update` target
will automatically be generated for your convenience.
