CONTAINER_ENGINE ?= docker

GOLANGCI_LINT_IMAGE := golangci/golangci-lint:latest

LINT_DOCKER_RUN := $(CONTAINER_ENGINE) run --rm -t \
	-v $(PWD):/app -w /app \
	--user $$(id -u):$$(id -g) \
	-v $$(go env GOCACHE):/.cache/go-build -e GOCACHE=/.cache/go-build \
	-v $$(go env GOMODCACHE):/.cache/mod -e GOMODCACHE=/.cache/mod \
	-v ~/.cache/golangci-lint:/.cache/golangci-lint -e GOLANGCI_LINT_CACHE=/.cache/golangci-lint \
	$(GOLANGCI_LINT_IMAGE)

.PHONY: format lint test build dogfood all

format:
	$(LINT_DOCKER_RUN) golangci-lint fmt

lint: format
	$(LINT_DOCKER_RUN) golangci-lint run --fix

test:
	go test -race -count=1 ./...

build:
	go build -o bin/gosigfmt ./cmd/gosigfmt

dogfood: build
	./bin/gosigfmt -l ./...

all: lint test build dogfood
