## FastTrackML
## For best results, run these make targets inside the devcontainer

#
# Tests targets.
#
.PHONY: tests-run-unit
tests-run-unit: ## run unit tests.
	@echo ">>> Running unit tests."
	go test -v ./...

.PHONY: tests-run-integration
tests-run-integration: ## run integration tests.
	@echo ">>> Running integration tests."
	go test -v -p 1 -tags="integration" ./tests/integration/...

PHONY: tests-mlflow
tests-mlflow: build
	@echo ">>> Running MLFlow integration tests."
	tests/mlflow/test.sh

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

#
# Build targets
# 
PHONY: clean
clean:
	@echo ">>> Cleaning node and go build artifacts."
	rm -Rf pkg/ui/aim/embed/repo
	rm -Rf pkg/ui/mlflow/embed/repo
	rm -Rf bin/fasttrack

PHONY: build
build:
	@echo ">>> Building go and node components."
	pkg/ui/aim/embed/build.sh
	pkg/ui/mlflow/embed/build.sh
	go build -o bin/fasttrack main.go

PHONY: run
run: build
	@echo ">>> Running the FasttrackML server."
	bin/fasttrack server
