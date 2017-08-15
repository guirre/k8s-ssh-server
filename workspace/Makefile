#!/usr/bin/make -f

# This allows us to have a GOPATH per peroject.
# Ensuring that we are all building with the same dependencies.
GOPATH := $(shell pwd):$(shell pwd)/vendor
export GOPATH
export CGO_ENABLED=0

NAME=k8s-ssh-server
PACKAGE=github.com/previousnext/$(NAME)

# Build binaries for linux/amd64 and darwin/amd64
build:
	gox -os='linux darwin' -arch='amd64' -output='bin/$(NAME)-gh-sync_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PACKAGE)/github-sync
	gox -os='linux darwin' -arch='amd64' -output='bin/$(NAME)-cli_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PACKAGE)/cli
	gox -os='linux darwin' -arch='amd64' -output='bin/$(NAME)-server_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PACKAGE)/server

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
