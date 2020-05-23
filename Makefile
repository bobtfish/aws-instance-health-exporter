IMAGE 		?= aws-instance-health-exporter
VERSION 	= $(shell git describe --always --tags --dirty)
GO_PACKAGES = $(shell go list ./... | grep -v /vendor/)

all: format build test

test:
	@go test $(GO_PACKAGES)

format:
	@echo ">> formatting code"
	@go fmt $(GO_PACKAGES)

build:
	@go build

docker:
	@docker build \
		--build-arg SOURCE_COMMIT="$(VERSION)" \
		-t $(IMAGE):$(VERSION) \
		.
	docker tag $(IMAGE):$(VERSION) $(IMAGE):latest

version:
	echo $(DOCKER_IMAGE_TAG)

.PHONY: all format build test docker
