# New function from git

It's possible to use Git repo as source for 
application/lambda.

Requirements:

* read-only access for the key (could be found in ui in Settings)
* manifest.json in repo

How to:

1. Add public key as allowed. 
For github: repo -> settings -> deploy keys
2. Put remote origin to field in repo and push {create from git} in UI in Dashboard
