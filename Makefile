#!/usr/bin/make -f

VERSION=$(shell git describe --tags --always)

release: build push

build:
	docker build -f dockerfiles/server/Dockerfile -t previousnext/k8s-ssh-server:${VERSION} .
	docker build -f dockerfiles/github/Dockerfile -t previousnext/k8s-ssh-github:${VERSION} .

push:
	docker push previousnext/k8s-ssh-server:${VERSION}
	docker push previousnext/k8s-ssh-github:${VERSION}

.PHONY: release build push
