---
layout: default
title: Git repo
parent: Usage
nav_order: 3
---
# New function from git

It's possible to use Git repo as source for 
application/lambda.

Requirements:

* read-only access for the key (could be found in UI in the Settings)
* manifest.json in repo

How to:

1. Add public key as allowed. 
For github: repo -> settings -> deploy keys
2. Put the remote origin URL into the field `Git repository` and push {create from git} in UI in Dashboard
