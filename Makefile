SHELL = /bin/bash
BINARY := agones-mc
IMAGE := saulmaldonado/$(BINARY)
COMMIT := $(shell git rev-parse --short HEAD)
VERSION := $(shell set -o pipefail; git describe --exact-match --tags HEAD 2> /dev/null | cut -c 2- || echo ${COMMIT})
BUILD_FLAGS ?= -v
ARCH ?= amd64
GOOGLE_APPLICATION_CREDENTIALS := $(HOME)/.config/gcloud/application_default_credentials.json
NAME := mc-server

-include .env
export

.PHONY: build build.docker build.docker-compose.monitor build.docker-compose.backup

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o build/$(BINARY) $(BUILD_FLAGS) .

build.docker:
	docker build --rm --tag $(IMAGE):$(VERSION) --build-arg VERSION=$(VERSION) --build-arg ARCH=$(ARCH) .

build.docker-compose.monitor:
	docker-compose -f monitor.docker-compose.yml build

build.docker-compose.backup:
	docker-compose -f backup.docker-compose.yml build

docker-compose.monitor:
	docker-compose -f monitor.docker-compose.yml up

docker-compose.backup:
	docker-compose -f backup.docker-compose.yml up

docker-compose.load:
	docker-compose -f load.docker-compose.yml up

docker-compose.fileserver:
	docker-compose -f fileserver.docker-compose.yml up

docker-compose.bedrock.backup:
	docker-compose -f backup.bedrock.docker-compose.yml up

docker-compose.bedrock.load:
	docker-compose -f load.bedrock.docker-compose.yml up

docker-compose.bedrock.fileserver:
	docker-compose -f fileserver.bedrock.docker-compose.yml up

clean:
	@rm -rf build

clean.docker: stop delete-containers delete-images

clean.docker-compose.backup:
	docker-compose -f backup.docker-compose.yml rm

clean.docker-compose.monitor:
	docker-compose -f monitor.docker-compose.yml rm

clean.docker-compose.load:
	docker-compose -f load.docker-compose.yml rm

stop:
	-docker container stop $(shell docker container ls -q --filter name=$(BINARY))

delete-containers:
	-docker rm $(shell docker ps -a -q --filter name=$(BINARY))

delete-images:
	-docker rmi $(shell docker images -q $(IMAGE)) -f
