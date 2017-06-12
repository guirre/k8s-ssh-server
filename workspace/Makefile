#!/usr/bin/make -f

server:
	./hack/build.sh linux server ssh-server github.com/previousnext/server

cli:
	./hack/build.sh linux cli ssh-cli github.com/previousnext/cli

github-sync:
	./hack/build.sh linux server github-sync github.com/previousnext/github-sync

.PHONY: server cli github-sync