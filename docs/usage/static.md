---
layout: default
title: Static files
parent: Usage
nav_order: 6
---
# Static files

Static files can be served by `GET` request from the specified folder.
Using static serving together with lambda allows you to create dynamic service with UI (like blog or comments).

To enable the feature, set the field `static` to relative path to the directory with static files
 (must be subfolder of lambda directory)  in [manifest](../../usage/manifest). 

If the feature is enabled the GET and HEAD methods will not be available for the handler (lambda).
Same security restrictions are applied to static files as to lambdas (security checks performed before file handling).

UI:

1. Click on already created app
2. Select mapping, enter static directory name relative to the application root, or clean it to remove
3. Click save

Related examples:

* [blog](../../examples/blog)
