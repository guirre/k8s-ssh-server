#!/usr/bin/make -f

build:
	./hack/build.sh linux server ssh-server github.com/previousnext/server
	./hack/build.sh linux cli ssh-cli github.com/previousnext/cli

docker:
	docker build -t previousnext/k8s-ssh-server .

.PHONY: build docker
