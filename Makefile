SHELL := /bin/bash

CURR_TIME = $(shell date +%s)

IMG_TAG_VERSION ?= latest

# So user will do make CONTROLLER_IMG_SUFFIX="/pr-1234"
IMG_SUFFIX ?=

include Makefile.components

KSCTL_AGENT_IMG ?= ghcr.io/ksctl/ksctl-agent${IMG_SUFFIX}:${IMG_TAG_VERSION}

KSCTL_STATE_IMPORTER_IMG ?= ghcr.io/ksctl/ksctl-stateimport${IMG_SUFFIX}:${IMG_TAG_VERSION}

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[36m  _             _   _ \n | |           | | | |\n | | _____  ___| |_| |\n | |/ / __|/ __| __| |\n |   <\\__ \\ (__| |_| |\n |_|\\_\\___/\\___|\\__|_| \033[0m\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Builder and Generator (Core)

.PHONY: gen-proto-agent
gen-proto-agent: check_protoc ## generate protobuf for ksctl agent
	@echo "generating new helpers"
	protoc --proto_path=api/agent/proto api/agent/proto/*.proto --go_out=api/gen/agent --go-grpc_out=api/gen/agent

.PHONY: check_protoc
check_protoc:
	@whereis protoc || (echo "Please install protoc" && exit 1)
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: docker-push-state-import
docker-push-state-import: ## Push docker image for ksctl state-import
	$(CONTAINER_TOOL) push ${KSCTL_STATE_IMPORTER_IMG}

.PHONY: docker-push-agent
docker-push-agent: ## Push docker image for ksctl agent
	$(CONTAINER_TOOL) push ${KSCTL_AGENT_IMG}

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx-agent
docker-buildx-agent: ## docker build agent
		- $(CONTAINER_TOOL) buildx create --name project-v3-builder
		$(CONTAINER_TOOL) buildx use project-v3-builder
		$(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg="GO_VERSION=1.22" --tag ${KSCTL_AGENT_IMG} -f build/agent/Dockerfile .
		- $(CONTAINER_TOOL) buildx rm project-v3-builder

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx-stateimport
docker-buildx-stateimport: ## docker build stateimport
		- $(CONTAINER_TOOL) buildx create --name project-v3-builder
		$(CONTAINER_TOOL) buildx use project-v3-builder
		$(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg="GO_VERSION=1.22" --tag ${KSCTL_STATE_IMPORTER_IMG} -f build/stateimport/Dockerfile .
		- $(CONTAINER_TOOL) buildx rm project-v3-builder


##@ Unit Tests (Core)
.PHONY: test_all
test_all:
	@go test -timeout 20m  -run 'All' -v tests/{runner_test.go,unit_test.go,integration_test.go}

.PHONY: unit_test
unit_test: ## Run unit tests
	@go test -timeout 20m -run 'Unit' -v tests/{runner_test.go,unit_test.go,integration_test.go}

.PHONY: integeration_test
integeration_test: ## Run integration tests
	@go test -timeout 20m -run 'Integration' -v tests/{runner_test.go,unit_test.go,integration_test.go}
