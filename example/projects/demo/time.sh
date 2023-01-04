#!/usr/bin/env sh
set -x
if [ "1$FORMAT" = "1" ]; then
  FORMAT="%Y-%m-%dT%H:%M:%S"
fi
date "+$FORMAT"