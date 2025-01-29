include mk/go.mk
# NOTE: rust.mk is included upstream

.PHONY: help
help:
	@echo "Available targets:"
	@echo
	@awk '/^## / { \
		helpMessage = substr($$0, 4); \
		getline; \
		if ($$1 == ".PHONY:") { \
			getline; \
		} \
		sub(/:.*/, "", $$1); \
		printf "  %-30s %s\n", $$1, helpMessage; \
	}' $(MAKEFILE_LIST)

## Install the stork-generate binary
.PHONY: install-stork-source-generator
install-stork-source-generator:
	@echo "Installing stork-source-generator..."
	@echo "Running: go build -o $(shell go env GOPATH)/bin/stork-source-generator ./apps/cmd/generate"
	@go build -o $(shell go env GOPATH)/bin/stork-source-generator ./apps/cmd/generate
	@go run ./apps/cmd/animate/main.go
	@echo "Successfully installed stork-source-generator. Run 'stork-source-generator help' to get started."

## Uninstall the stork-source-generator binary
.PHONY: uninstall-stork-source-generator
uninstall-stork-source-generator:
	@echo "Uninstalling stork-source-generator..."
	@echo "Running: rm -f $(shell go env GOPATH)/bin/stork-source-generator"
	@rm -f $(shell go env GOPATH)/bin/stork-source-generator
	@echo "Successfully uninstalled stork-source-generator"

## Install the stork-data-provider binary
.PHONY: install-stork-source-runner
install-stork-source-runner:
	@echo "Installing stork-source-runner..."
	@echo "Running: go build -o $(shell go env GOPATH)/bin/stork-source-runner ./apps/cmd/data_provider"
	@go build -o $(shell go env GOPATH)/bin/stork-source-runner ./apps/cmd/data_provider
	@echo "Successfully installed stork-source-runner. Run 'stork-source-runner help' to get started."

## Uninstall the stork-data-provider binary
.PHONY: uninstall-stork-source-runner
uninstall-stork-source-runner:
	@echo "Uninstalling stork-source-runner..."
	@rm -f $(shell go env GOPATH)/bin/stork-source-runner
	@echo "Successfully uninstalled stork-source-runner"

## Rebuild and reinstall the stork-data-provider binary
.PHONY: rebuild-stork-source-runner
rebuild-stork-source-runner: uninstall-stork-source-runner install-stork-source-runner
	@echo "Successfully rebuilt stork-source-runner"

## Start the data provider (rebuilds first)
.PHONY: start-stork-source-runner
start-stork-source-runner: rebuild-stork-source-runner
	@if [ -z "$(ARGS)" ]; then \
		echo "Error: Missing required arguments"; \
		echo "Usage: make start-stork-source-runner ARGS=\"-c <config-file-path> -o <output-address>\""; \
		exit 1; \
	fi
	@echo "Starting data provider with arguments: $(ARGS)"
	@stork-source-runner start $(ARGS)
