SHELL := /bin/bash

CURR_TIME = $(shell date +%s)

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\n\033[36m  _             _   _ \n | |           | | | |\n | | _____  ___| |_| |\n | |/ / __|/ __| __| |\n |   <\\__ \\ (__| |_| |\n |_|\\_\\___/\\___|\\__|_| \033[0m\n\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Tests
.PHONY: test_all
test_all:
	@go test -timeout 20m  -run 'All' -v tests/{runner_test.go,unit_test.go,integration_test.go}

##@ Unit Tests
.PHONY: unit_test
unit_test: ## Run unit tests
	@go test -timeout 20m -run 'Unit' -v tests/{runner_test.go,unit_test.go,integration_test.go}

##@ Integeration Tests
.PHONY: integeration_test
integeration_test: ## Run integration tests
	@go test -timeout 20m -run 'Integration' -v tests/{runner_test.go,unit_test.go,integration_test.go}
