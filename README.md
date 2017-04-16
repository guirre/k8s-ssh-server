Kubernetes: SSH Server
======================

A ThirdPartyResource SSH Server for connecting to containers on a Kubernetes cluster.

## Features

* Per namespace users eg. "namespace1" cannot connect to "namespace2".
* Window resizing
* Works with rsync

## Testing

```bash
$ kubectl create -f ssh-server

# Deploy this ssh server.
$ kubectl -n ssh-server create -f examples/ssh-server.yaml

# Create some containers to test shell access.
$ kubectl -n ssh-server create -f examples/deployment.yaml

# Make sure to update the authorized key value. 
$ kubectl -n ssh-server create -f examples/ssh-server.yaml

# Get a list of the pods, update the POD value in the ssh command.
$ kubectl -n ssh-server get pods
$ ssh ssh-server~POD~nginx~nick@localhost -p 2222
```

## Building

```bash
$ make
```
