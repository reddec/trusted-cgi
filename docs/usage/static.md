---
layout: default
title: Static files
parent: Usage
nav_order: 6
---
# Static files

Static files can be served by `GET` request from the specified folder.
Using static serving together with lambda allows you to create dynamic service with UI (like blog or comments).

To enable the feature, set in [manifest](../../usage/manifest) field `static` to relative path to the directory with static files
 (must be subfolder of lambda directory). 

If the feature enabled the GET and HEAD methods will not be available for handler (lambda).
Same security restrictions applied to static files as to lambda (security checks performed before file handling).

UI:

1. Click to already create app
2. Select mapping, enter static directory name, relative to the application root, or clean ir to remove
3. Click save

Related examples:

* [blog](../../examples/blog)