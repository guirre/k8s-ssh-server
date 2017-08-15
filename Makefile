#!/usr/bin/make -f

VERSION=$(shell git describe --tags --always)

release: build push

build:
	docker build -f dockerfiles/server/Dockerfile -t previousnext/k8s-ssh-server:${VERSION} .
	docker build -f dockerfiles/github-sync/Dockerfile -t previousnext/k8s-ssh-server-gh:${VERSION} .

push:
	docker push previousnext/k8s-ssh-server:${VERSION}
	docker push previousnext/k8s-ssh-server-gh:${VERSION}

.PHONY: release build push
