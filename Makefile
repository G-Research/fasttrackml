#
# Tests targets.
#
.PHONY: tests-run-unit
tests-run-unit: ## run unit tests.
	@echo ">>> Running unit tests."
	@env go test -v ./...

.PHONY: tests-run-integration
tests-run-integration: ## run integration tests.
	@echo ">>> Running integration tests."
	@env go test -v -p 1 -tags="integration" ./tests/integration/...


#
# Mockery targets.
#
.PHONY: mocks-clean
mocks-clean: ## cleans old mocks.
	find . -name "mock_*.go" -type f -print0 | xargs -0 /bin/rm -f

.PHONY: mocks-generate
mocks-generate: mocks-clean ## generate mock based on all project interfaces.
	mockery --all --dir "./pkg/api/mlflow" --inpackage --case underscore

.PHONY: help
help: ## display this help
	@ echo "Please use \`make <target>' where <target> is one of:"
	@ echo
	@ grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-10s\033[0m - %s\n", $$1, $$2}'
	@ echo

