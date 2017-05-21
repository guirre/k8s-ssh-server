#!/usr/bin/make -f

server:
	./hack/build.sh linux server ssh-server github.com/previousnext/server

cli:
	./hack/build.sh linux cli ssh-cli github.com/previousnext/cli

github-sync:
	./hack/build.sh linux cli github-sync github.com/previousnext/github-sync

docker:
	docker build -t previousnext/k8s-ssh-server .

.PHONY: build docker
