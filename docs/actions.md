# Actions

Actions is optional arbitrary commands defined in [Makefile](https://www.gnu.org/software/make/manual/make.html#Rule-Example) as targets and can be invoked
by UI, admin API or during template cloning operations.

Main purpose is to prepare environment or function out of general flow procedure (HTTP call): 
build binary, download dependencies, etc.

UI:
 
1. click to any created application
2. click to actions tab