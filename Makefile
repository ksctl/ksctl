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

.PHONY: docker-build-agent
docker-build-agent: ## docker build agent
	docker build \
		--file build/agent/Dockerfile \
		--build-arg="GO_VERSION=1.22" \
		--platform="linux/amd64" \
		--tag ${KSCTL_AGENT_IMG} .

.PHONY: docker-build-state-import
docker-build-state-import: ## docker build state importer
	docker build \
		--file build/stateimport/Dockerfile \
		--build-arg="GO_VERSION=1.22" \
		--platform="linux/amd64" \
		--tag ${KSCTL_STATE_IMPORTER_IMG} .

##@ Unit Tests (Core)
.PHONY: unit_test_all
unit_test_all:
	@go test -v tests/unit_test.go

##@ Mock Tests (Core)
.PHONY: mock_all
mock_all: golang-test ## All Mock tests
	@echo "Mock Test (integration)"
	cd test/ && \
		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_aws,testing_civo,testing_azure,testing_local -bench=. -benchtime=1x -cover -v

.PHONY: mock_civo_ha
mock_civo_ha: golang-test ## Civo HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_civo -bench=BenchmarkCivoTestingHA -benchtime=1x -cover -v

.PHONY: mock_civo_managed
mock_civo_managed: golang-test ## Civo managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_civo -bench=BenchmarkCivoTestingManaged -benchtime=1x -cover -v

.PHONY: mock_azure_managed
mock_azure_managed: golang-test ## Azure managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_azure -bench=BenchmarkAzureTestingManaged -benchtime=1x -cover -v

.PHONY: mock_azure_ha
mock_azure_ha: golang-test ## Azure HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_azure -bench=BenchmarkAzureTestingHA -benchtime=1x -cover -v

.PHONY: mock_aws_ha
mock_aws_ha: golang-test ## Aws HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_aws -bench=BenchmarkAwsTestingHA -benchtime=1x -cover -v

.PHONY: mock_aws_managed
mock_aws_managed: golang-test ## Aws managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_aws -bench=BenchmarkAwsTestingManaged -benchtime=1x -cover -v

.PHONY: mock_local_managed
mock_local_managed: golang-test ## Local managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -tags testing_local -bench=BenchmarkLocalTestingManaged -benchtime=1x -cover -v


##@ Complete Testing (Core)
.PHONY: test-core
test-core: unit_test_all mock_all ## do both unit and integration test
	@echo "Done All tests"



.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	@echo -e "\n\033[36mRunning for Ksctl (Core)\033[0m\n" && \
		$(GOLANGCI_LINT) run --timeout 10m && echo -e "\n=========\n\033[91m✔ PASSED\033[0m\n=========\n" || echo -e "\n=========\n\033[91m✖ FAILED\033[0m\n=========\n"
	@echo -e "\n\033[36mRunning for Ksctl (Agent)\033[0m" && \
		cd ksctl-components/agent && \
		$(GOLANGCI_LINT) run --timeout 10m && echo -e "\n=========\n\033[91m✔ PASSED\033[0m\n=========\n" || echo -e "\n=========\n\033[91m✖ FAILED\033[0m\n=========\n"
	@echo -e "\n\033[36mRunning for Ksctl (StateImport)\033[0m" && \
		cd ksctl-components/stateimport && \
		$(GOLANGCI_LINT) run --timeout 10m && echo -e "\n=========\n\033[91m✔ PASSED\033[0m\n=========\n" || echo -e "\n=========\n\033[91m✖ FAILED\033[0m\n=========\n"
	@echo -e "\n\033[36mRunning for Ksctl Controllers (Application)\033[0m" && \
		make lint-controller CONTROLLER=application && echo -e "\n=========\n\033[91m✔ PASSED\033[0m\n=========\n" || echo -e "\n=========\n\033[91m✖ FAILED\033[0m\n=========\n"

.PHONY: test
test: lint
	make test-core
	@echo -e "\n\033[36mTesting in ksctl-components\033[0m\n"
	make test-controller CONTROLLER=application
