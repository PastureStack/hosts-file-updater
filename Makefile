VERSION ?= $(shell ./scripts/version)
IMAGE_NAME ?= ghcr.io/pasturestack/hosts-file-updater
GO_VERSION ?= 1.26.5
UBUNTU_VERSION ?= 26.04
DOCKER_BUILD_NETWORK ?= default

ci: test image

test:
	docker build --target test \
		--network $(DOCKER_BUILD_NETWORK) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg UBUNTU_VERSION=$(UBUNTU_VERSION) \
		-t $(IMAGE_NAME)-test:$(VERSION) .

image:
	docker build --target runtime \
		--network $(DOCKER_BUILD_NETWORK) \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg UBUNTU_VERSION=$(UBUNTU_VERSION) \
		--build-arg VERSION=$(VERSION) \
		-t $(IMAGE_NAME):$(VERSION) .

push: image
	docker push $(IMAGE_NAME):$(VERSION)

.DEFAULT_GOAL := ci

.PHONY: ci test image push
