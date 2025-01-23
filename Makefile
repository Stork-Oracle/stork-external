## Install the stork-data-provider binary
install-data-provider-cli:
	@echo "Installing stork-data-provider..."
	@go build -o $(shell go env GOPATH)/bin/stork-data-provider ./apps/cmd/data_provider
	@echo "Successfully installed stork-data-provider. Run 'stork-data-provider help' to get started."

## Uninstall the stork-data-provider binary
uninstall-data-provider-cli:
	@echo "Uninstalling stork-data-provider..."
	@rm -f $(shell go env GOPATH)/bin/stork-data-provider $(shell go env GOPATH)/bin/data_provider
	@echo "Successfully uninstalled stork-data-provider"

# Include other makefiles after the CLI targets
include mk/go.mk
# NOTE: rust.mk is included upstream

.PHONY: help install-data-provider-cli uninstall-data-provider-cli
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
