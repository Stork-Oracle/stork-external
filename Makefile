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
