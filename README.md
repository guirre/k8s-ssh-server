Kubernetes: SSH Server
======================

A ThirdPartyResource SSH Server for connecting to containers on a Kubernetes cluster.

## Features

* Per namespace users eg. "namespace1" cannot connect to "namespace2".
* Window resizing
* Works with rsync

## Release

By default `make release` will tag images as "latest".

To pick a version (**and you should!**) run the following command which will:

* Make new images
* Push them to the Docker Hub

```bash
make release VERSION=0.0.1
```