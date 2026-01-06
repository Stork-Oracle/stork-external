include mk/rust.mk

LIB_DIR ?= $(RUST_LIB_DIR)
include mk/misc.mk

CGO_LDFLAGS := "-L$(RUST_LIB_DIR)"
GOBIN := $(PWD)/.bin
GO := GOBIN=$(GOBIN) CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS) LD_LIBRARY_PATH=$(RUST_LIB_DIR):$(LD_LIBRARY_PATH) GOPRIVATE="*" GIT_TERMINAL_PROMPT=0 go

$(GOBIN):
	@mkdir -p $(GOBIN)

MOCKERY_MAJOR_VERSION := 3
MOCKERY_VERSION := v3.5.0

$(GOBIN)/mockery: $(GOBIN)
	@echo "Installing mockery $(MOCKERY_VERSION)..."
	@$(GO) install github.com/vektra/mockery/v$(MOCKERY_MAJOR_VERSION)@$(MOCKERY_VERSION)

.PHONY: mocks
## Generate mocks for the project using mockery
mocks: $(GOBIN)/mockery
	@echo "Generating mocks..."
	@$(GOBIN)/mockery --config .mockery.yaml

# TODO: Add race checker
.PHONY: test
## Run all Go tests
test: signer_ffi fuel_ffi
	@$(GO) test -v ./...

.PHONY: integration-test
## Run all Go integration tests
integration-test: signer_ffi fuel_ffi
	@echo "Running Go integration tests..."
	@set -e; \
	for pkg in $$($(GO) list ./... | grep -v "/integration$$" | grep -v "apps/first_party_pusher"); do \
	    $(GO) test -v -tags integration $$pkg; \
	done

.PHONY: first-party-integration-test
## Run all Go integration tests
first-party-integration-test: signer_ffi
	@echo "Running Go integration tests..."
	@set -e; \
	$(GO) test -v -tags integration ./apps/first_party_pusher/pkg/evm/...


# Individual Go Targets
chain_pusher: $(if $(NO_RUST),,signer_ffi fuel_ffi) wasmvm
	@echo "Installing chain pusher..."
	@$(GO) install -v ./apps/chain_pusher

publisher_agent: $(if $(NO_RUST),,signer_ffi)
	@echo "Installing publisher agent..."
	@$(GO) install -v ./apps/publisher_agent

data_provider:
	@echo "Installing data provider..."
	@$(GO) install -v ./apps/data_provider

generate:
	@echo "Installing generate..."
	@$(GO) install -v ./utils/generate

first_party_pusher: $(if $(NO_RUST),,signer_ffi)
	@echo "Installing first party pusher..."
	@$(GO) install -v ./apps/first_party_pusher

.PHONY: install
## Aggregate target to install all Go binaries
install: $(if $(NO_RUST),,signer_ffi fuel_ffi) chain_pusher publisher_agent data_provider generate wasmvm first_party_pusher
	@echo "All Go binaries have been installed to $(GOBIN) successfully."

.PHONY: clean
## Clean up the project
clean: clean-rust clean-misc
	@rm -rf $(GOBIN)
	@$(GO) clean -cache -testcache

# pass in a target to run-local to run a specific binary
run-local: signer_ffi fuel_ffi wasmvm
	@$(GO) run ./apps/$(target) $(args)

# Lint Go code using golangci-lint
.PHONY: lint-go
lint-go:
	@echo "Linting Go code..."
	@golangci-lint run

# Format Go code using golangci-lint formatters
.PHONY: format-go
format-go:
	@echo "Formatting Go code..."
	@golangci-lint fmt
