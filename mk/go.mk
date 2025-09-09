include mk/rust.mk

CGO_LDFLAGS := "-L$(RUST_LIB_DIR)"
GOBIN := $(PWD)/.bin
GO := GOBIN=$(GOBIN) CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS) LD_LIBRARY_PATH=$(RUST_LIB_DIR):$(LD_LIBRARY_PATH) GOPRIVATE="*" GIT_TERMINAL_PROMPT=0 go

$(GOBIN):
	@mkdir -p $(GOBIN)

$(GOBIN)/mockery: $(GOBIN)
	@echo "Installing mockery..."
	@$(GO) install github.com/vektra/mockery/v2@v2.46.3

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

.PHONY: install-cosmwasm-libs
install-cosmwasm-libs:
	@mkdir -p .lib
	@if [ ! -f ".lib/libwasmvm.$(shell uname -m | sed 's/x86_64/x86_64/;s/aarch64/aarch64/').so" ]; then \
		echo "Installing CosmWasm libraries..."; \
		curl -L https://github.com/CosmWasm/wasmvm/releases/download/v2.2.1/libwasmvm.$(shell uname -m | sed 's/x86_64/x86_64/;s/aarch64/aarch64/').so -o .lib/libwasmvm.$(shell uname -m | sed 's/x86_64/x86_64/;s/aarch64/aarch64/').so; \
		echo "Successfully installed CosmWasm libraries to .lib/"; \
	else \
		echo "CosmWasm libraries already installed"; \
	fi

# Individual Go Targets
chain_pusher: signer_ffi fuel_ffi
	@$(GO) install -v ./apps/chain_pusher/cmd

publisher_agent: signer_ffi
	@$(GO) install -v ./apps/publisher_agent/cmd

data_provider: 
	@$(GO) install -v ./apps/data_provider/cmd

generate: 
	@$(GO) install -v ./tools/generate/cmd

.PHONY: install
## Aggregate target to install all Go binaries	
install: chain_pusher publisher_agent data_provider generate install-cosmwasm-libs
	@echo "All Go binaries have been installed successfully."

.PHONY: clean
## Clean up the project
clean: clean-rust
	@rm -rf $(GOBIN)
	@$(GO) clean -cache -testcache

# pass in a target to run-local to run a specific binary
run-local: signer_ffi fuel_ffi install-cosmwasm-libs
	@$(GO) run ./apps/$(target)/cmd $(args)

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
