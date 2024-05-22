SHELL := /bin/bash

CURR_TIME = $(shell date +%s)

IMG_TAG_VERSION ?= latest

include Makefile.components

KSCTL_AGENT_IMG ?= ghcr.io/ksctl/ksctl-agent:${IMG_TAG_VERSION}

KSCTL_STATE_IMPORTER_IMG ?= ghcr.io/ksctl/ksctl-stateimport:${IMG_TAG_VERSION}

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[36m  _             _   _ \n | |           | | | |\n | | _____  ___| |_| |\n | |/ / __|/ __| __| |\n |   <\\__ \\ (__| |_| |\n |_|\\_\\___/\\___|\\__|_| \033[0m\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Builder and Generator (Core)

.PHONY: gen-proto-agent
gen-proto-agent: ## generate protobuf for ksctl agent
	@echo "generating new helpers"
	protoc --proto_path=api/agent/proto api/agent/proto/*.proto --go_out=api/gen/agent --go-grpc_out=api/gen/agent

.PHONY: docker-push-state-import
docker-push-state-import: ## Push docker image for ksctl state-import
	$(CONTAINER_TOOL) push ${KSCTL_STATE_IMPORTER_IMG}

.PHONY: docker-push-agent
docker-push-agent: ## Push docker image for ksctl agent
	$(CONTAINER_TOOL) push ${KSCTL_AGENT_IMG}

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx-agent
docker-buildx-agent: ## docker build agent
		- $(CONTAINER_TOOL) buildx create --name project-v3-builder || exit 1
		$(CONTAINER_TOOL) buildx use project-v3-builder || exit 1
		- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg="GO_VERSION=1.22" --tag ${KSCTL_AGENT_IMG} -f build/agent/Dockerfile . || exit 1
		 - $(CONTAINER_TOOL) buildx rm project-v3-builder || exit 1

PLATFORMS ?= linux/arm64,linux/amd64
.PHONY: docker-buildx-stateimport
docker-buildx-stateimport: ## docker build stateimport
		- $(CONTAINER_TOOL) buildx create --name project-v3-builder || exit 1
		$(CONTAINER_TOOL) buildx use project-v3-builder || exit 1
		- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --build-arg="GO_VERSION=1.22" --tag ${KSCTL_STATE_IMPORTER_IMG} -f build/stateimport/Dockerfile . || exit 1
		 - $(CONTAINER_TOOL) buildx rm project-v3-builder || exit 1

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
unit_test_all: golang-test ## all unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-api.sh $(GO_TEST_COLOR)

.PHONY: unit_test_utils
unit_test_utils: golang-test ## utils unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-utils.sh $(GO_TEST_COLOR)

.PHONY: unit_test_logger
unit_test_logger: golang-test ## logger unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-logger.sh $(GO_TEST_COLOR)

.PHONY: unit_test_civo
unit_test_civo: golang-test ## civo unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-civo.sh $(GO_TEST_COLOR)

.PHONY: unit_test_local
unit_test_local: golang-test ## local unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-local.sh $(GO_TEST_COLOR)

.PHONY: unit_test_kubernetes_apps
unit_test_kubernetes_apps: golang-test ## azure unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-kubernetes.sh $(GO_TEST_COLOR)

.PHONY: unit_test_azure
unit_test_azure: golang-test ## azure unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-azure.sh $(GO_TEST_COLOR)

.PHONY: unit_test_aws
unit_test_aws: golang-test ## aws unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-aws.sh $(GO_TEST_COLOR)

.PHONY: unit_test_k3s
unit_test_k3s: golang-test ## k3s unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-k3s.sh $(GO_TEST_COLOR)

.PHONY: unit_test_kubeadm
unit_test_kubeadm: golang-test ## kubeadm unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-kubeadm.sh $(GO_TEST_COLOR)


.PHONY: unit_test_bootstrap
unit_test_bootstrap: golang-test ## bootstrap unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-bootstrap.sh $(GO_TEST_COLOR)


.PHONY: unit_test_kubernetes-store
unit_test_kubernetes-store: golang-test ## kubernetes-store unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-kubernetes-store.sh $(GO_TEST_COLOR)

.PHONY: unit_test_local-store
unit_test_local-store: golang-test ## local-store unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-local-store.sh $(GO_TEST_COLOR)


.PHONY: unit_test_mongodb-store
unit_test_mongodb-store: golang-test ## mongodb-store unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-mongodb-store.sh $(GO_TEST_COLOR)

.PHONY: unit_test_ksctl_agent
unit_test_ksctl_agent: golang-test ## ksctl-agent unit test case
	@echo "Unit Tests"
	cd scripts/ && \
		/bin/bash test-ksctl-agent.sh $(GO_TEST_COLOR)

##@ Mock Tests (Core)
.PHONY: mock_all
mock_all: golang-test ## All Mock tests
	@echo "Mock Test (integration)"
	cd test/ && \
		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=. -benchtime=1x -cover -v

.PHONY: mock-civo-ha
mock_civo_ha: golang-test ## Civo HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkCivoTestingHA -benchtime=1x -cover -v

.PHONY: mock-civo-managed
mock_civo_managed: golang-test ## Civo managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkCivoTestingManaged -benchtime=1x -cover -v

.PHONY: mock-azure-managed
mock_azure_managed: golang-test ## Azure managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkAzureTestingManaged -benchtime=1x -cover -v

.PHONY: mock-azure-ha
mock_azure_ha: golang-test ## Azure HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkAzureTestingHA -benchtime=1x -cover -v

.PHONY: mock-aws-ha
mock_aws_ha: golang-test ## Aws HA mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkAwsTestingHA -benchtime=1x -cover -v


.PHONY: mock-local-managed
mock_local_managed: golang-test ## Local managed mock test
	cd test/ && \
 		GOTEST_PALETTE="red,yellow,green" $(GO_TEST_COLOR) -bench=BenchmarkLocalTestingManaged -benchtime=1x -cover -v


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
