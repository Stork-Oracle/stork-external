include mk/go.mk
include mk/lint.mk
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

## Install the stork-source-generate binary
.PHONY: install-stork-generate
install-stork-generate:
	@echo "Installing stork-generate..."
	@echo "Running: go build -o $(shell go env GOPATH)/bin/stork-generate ./apps/cmd/generate"
	@go build -o $(shell go env GOPATH)/bin/stork-generate ./apps/cmd/generate
	@./apps/scripts/animate.sh
	@echo "Successfully installed stork-generate. Run 'stork-generate help' to get started."

## Uninstall the stork-generate binary
.PHONY: uninstall-stork-generate
uninstall-stork-generate:
	@echo "Uninstalling stork-generate..."
	@echo "Running: rm -f $(shell go env GOPATH)/bin/stork-generate"
	@rm -f $(shell go env GOPATH)/bin/stork-generate
	@echo "Successfully uninstalled stork-generate"

## Install the stork-source-runner binary
.PHONY: install-data-provider
install-data-provider:
	@echo "Installing data-provider..."
	@echo "Running: go build -o $(shell go env GOPATH)/bin/data-provider ./apps/cmd/data_provider"
	@go build -o $(shell go env GOPATH)/bin/data-provider ./apps/cmd/data_provider
	@echo "Successfully installed data-provider. Run 'data-provider help' to get started."

## Uninstall the stork-source-runner binary
.PHONY: uninstall-data-provider
uninstall-data-provider:
	@echo "Uninstalling data-provider..."
	@rm -f $(shell go env GOPATH)/bin/data-provider
	@echo "Successfully uninstalled data-provider"

## Rebuild and reinstall the stork-source-runner binary
.PHONY: rebuild-data-provider
rebuild-data-provider: uninstall-data-provider install-data-provider
	@echo "Successfully rebuilt data-provider"

## Start the source runner (rebuilds first)
.PHONY: start-data-provider
start-data-provider: rebuild-data-provider
	@if [ -z "$(ARGS)" ]; then \
		echo "Error: Missing required arguments"; \
		echo "Usage: make start-data-provider ARGS=\"-c <config-file-path> -o <output-address>\""; \
		exit 1; \
	fi
	@echo "Starting data provider with arguments: $(ARGS)"
	@data-provider start $(ARGS)
