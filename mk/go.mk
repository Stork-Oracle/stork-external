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
test: $(LIBSTORK) $(LIBFUEL_FFI)
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

.PHONY: install
## Aggregate target to install all Go binaries
install: $(LIBSTORK) install-cosmwasm-libs
	@$(GO) install -v ./apps/cmd/...
	@echo "All Go binaries have been installed successfully."

.PHONY: clean
## Clean up the project
clean: clean-rust
	@rm -rf $(GOBIN)
	@$(GO) clean -cache -testcache

# pass in a target to run-local to run a specific binary
run-local: $(LIBSTORK)
	@$(GO) run ./apps/cmd/$(target) $(args)
