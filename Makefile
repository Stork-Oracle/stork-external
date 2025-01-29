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

## Install the stork-data-provider binary
.PHONY: install-data-provider-cli
install-data-provider-cli:
	@echo "Installing stork-data-provider..."
	@echo "Running: go build -o $(shell go env GOPATH)/bin/stork-data-provider ./apps/cmd/data_provider"
	@go build -o $(shell go env GOPATH)/bin/stork-data-provider ./apps/cmd/data_provider
	@echo "Successfully installed stork-data-provider. Run 'stork-data-provider help' to get started."

## Uninstall the stork-data-provider binary
.PHONY: uninstall-data-provider-cli
uninstall-data-provider-cli:
	@echo "Uninstalling stork-data-provider..."
	@rm -f $(shell go env GOPATH)/bin/stork-data-provider $(shell go env GOPATH)/bin/data_provider
	@echo "Successfully uninstalled stork-data-provider"

## Rebuild and reinstall the stork-data-provider binary
.PHONY: rebuild-data-provider-cli
rebuild-data-provider-cli: uninstall-data-provider-cli install-data-provider-cli
	@echo "Successfully rebuilt stork-data-provider"

## Start the data provider (rebuilds first)
.PHONY: start-data-provider
start-data-provider: rebuild-data-provider-cli
	@if [ -z "$(ARGS)" ]; then \
		echo "Error: Missing required arguments"; \
		echo "Usage: make start-data-provider ARGS=\"-c <config-file-path> -o <output-address>\""; \
		exit 1; \
	fi
	@echo "Starting data provider with arguments: $(ARGS)"
	@stork-data-provider start $(ARGS)
