SHELL = /bin/bash
BINARY := agones-mc
IMAGE := saulmaldonado/$(BINARY)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell set -o pipefail; git describe --exact-match --tags HEAD 2> /dev/null | cut -c 2- || echo ${COMMIT})
BUILD_FLAGS ?= -v
ARCH ?= amd64

.PHONY: build build.docker

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o build/$(BINARY) $(BUILD_FLAGS) .

build.docker:
	docker build --rm --tag $(IMAGE):$(VERSION) --build-arg VERSION=$(VERSION) --build-arg ARCH=$(ARCH) .

clean:
	@rm -rf build

clean.docker: stop delete-containers delete-images

stop:
	-docker container stop $(shell docker container ls -q --filter name=$(BINARY))

delete-containers:
	-docker rm $(shell docker ps -a -q --filter name=$(BINARY))

delete-images:
	-docker rmi $(shell docker images -q $(IMAGE)) -f
