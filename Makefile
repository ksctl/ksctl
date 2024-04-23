SHELL := /bin/bash

CURR_TIME = $(shell date +%s)

include Makefile.components

KSCTL_AGENT_IMG ?= ghcr.io/ksctl/ksctl-agent:latest


.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[36m  _             _   _ \n | |           | | | |\n | | _____  ___| |_| |\n | |/ / __|/ __| __| |\n |   <\\__ \\ (__| |_| |\n |_|\\_\\___/\\___|\\__|_| \033[0m\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Builder and Generator (Core)

.PHONY: gen-proto-agent
gen-proto-agent: ## generate protobuf for ksctl agent
	@echo "generating new helpers"
	protoc --proto_path=api/agent/proto api/agent/proto/*.proto --go_out=api/gen/agent --go-grpc_out=api/gen/agent


.PHONY: docker-push-agent
docker-push-agent: ## Push docker image for ksctl agent
	$(CONTAINER_TOOL) push ${KSCTL_AGENT_IMG}

PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx-agent
docker-buildx-agent: ## docker build agent
		sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' build/agent/Dockerfile > build/agent/Dockerfile.cross
		- $(CONTAINER_TOOL) buildx create --name project-v3-builder
		$(CONTAINER_TOOL) buildx use project-v3-builder
		- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg="GO_VERSION=1.21" --tag ${KSCTL_AGENT_IMG} -f build/agent/Dockerfile.cross .
		 - $(CONTAINER_TOOL) buildx rm project-v3-builder
		 rm build/agent/Dockerfile.cross

.PHONY: docker-build-agent
docker-build-agent: ## docker build agent
	docker build \
		--file build/agent/Dockerfile \
		--build-arg="GO_VERSION=1.21" \
		--tag ${KSCTL_AGENT_IMG} .


##@ Unit Tests (Core)
.PHONY: unit-all
unit_test_api: ## all unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		chmod u+x test-api.sh && \
		./test-api.sh

##@ Mock Tests (Core)
.PHONY: mock-all
mock_test: ## All Mock tests
	@echo "Mock Test (integration)"
	cd test/ && \
		go test -bench=. -benchtime=1x -cover -v

.PHONY: mock-civo-ha
mock_civo_ha: ## Civo HA mock test
	cd test/ && \
 		go test -bench=BenchmarkCivoTestingHA -benchtime=1x -cover -v

.PHONY: mock-civo-managed
mock_civo_managed: ## Civo managed mock test
	cd test/ && \
 		go test -bench=BenchmarkCivoTestingManaged -benchtime=1x -cover -v

.PHONY: mock-azure-managed
mock_azure_managed: ## Azure managed mock test
	cd test/ && \
 		go test -bench=BenchmarkAzureTestingManaged -benchtime=1x -cover -v

.PHONY: mock-azure-ha
mock_azure_ha: ## Azure HA mock test
	cd test/ && \
 		go test -bench=BenchmarkAzureTestingHA -benchtime=1x -cover -v

.PHONY: mock-local-managed
mock_local_managed: ## Local managed mock test
	cd test/ && \
 		go test -bench=BenchmarkLocalTestingManaged -benchtime=1x -cover -v


##@ Complete Testing (Core)
.PHONY: test-all
test: unit_test_api mock_test ## do both unit and integration test
	@echo "Done All tests"
