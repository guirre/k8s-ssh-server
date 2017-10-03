#!/usr/bin/make -f

# This allows us to have a GOPATH per peroject.
# Ensuring that we are all building with the same dependencies.
GOPATH := $(shell pwd):$(shell pwd)/vendor
export GOPATH
export CGO_ENABLED=0

PACKAGE=github.com/previousnext/k8s-ssh

# Build binaries for linux/amd64 and darwin/amd64
build: server githhub

server:
	gox -os='linux' -arch='amd64' -output='bin/server_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PACKAGE)/server

github:
	gox -os='linux' -arch='amd64' -output='bin/github_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PACKAGE)/provider/github

# Run all lint checking with exit codes for CI
# We have to skip "client" because K8s TPRs require camel case eg. SshUser.
lint:
	golint -set_exit_status $(PACKAGE)/cli/...
	golint -set_exit_status $(PACKAGE)/server/...
	golint -set_exit_status $(PACKAGE)/log/...
	golint -set_exit_status $(PACKAGE)/github-sync/...

# Run tests with coverage reporting
test:
	go test -cover $(PACKAGE)/...
