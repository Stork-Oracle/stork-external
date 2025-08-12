.PHONY: lint-links

### Checks for broken links in all markdown files
lint-links:
	@if [ -z "$(shell git diff-index --quiet HEAD --)" ]; then \
		echo "Warning: you have uncommitted changes, this may return false positives for externalized links to this repo."; \
	fi
	@CURRENT_BRANCH=$(shell git rev-parse HEAD) linkspector check -c .linkspector.yml 
