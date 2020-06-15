---
layout: default
title: Templates
nav_order: 5
has_children: true
---

# Templates

Templates are a special kind of JSON-based files that defines a structure and base files
for a new lambdas/functions/apps.

Templates could be embedded and external. External templates folder by-default should be located in a
working directory as `.templates` and could be changed by `--templates` flag of `TEMPLATES` environment.

Name file will be a name of template (except .json).

Templates could define files, [manifest](../usage/manifest), [actions](../usage/actions) to invoke after clone and required checks.

Minimal working template file:

```json
{
  "manifest": {
    "run": ["echo", "hello world"]
  }
}
```

Structure:

* **description** (optional, string): short description of template
* **manifest** (required, [Manifest](../usage/manifest)): manifest definition for a new lambda
* **post_clone** (optional, string): [action](../usage/actions) to invoke after clone
* **checks** (optional, array of array of string): list of commands to invoke to check template availability (see example below)
* **files** (optional, map of string to string): files and content in a new lambda


If at least one check failed - template will be disabled.

Example check to ensure that template will be available only if python3 and pip3 installed:

```json
{
  "checks": [
    ["which", "python3"],
    ["which", "pip3"]
  ]
}
```

## Embedded

Most embeddable templates will be available in Docker image or via installing debian package (with
recommended packages - installed automatically by default)

List of embedded templates below: