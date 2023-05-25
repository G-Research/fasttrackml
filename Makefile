## FastTrackML
## For best results, run these make targets inside the devcontainer

#
# Tests targets.
#
.PHONY: test-go-unit
test-go-unit: ## run go unit tests.
	@echo ">>> Running unit tests."
	go test -v ./...

.PHONY: test-go-integration
test-go-integration: ## run go integration tests.
	@echo ">>> Running integration tests."
	go test -v -p 1 -tags="integration" ./tests/integration/...

PHONY: test-python-integration
test-python-integration: build ## run the MLFlow python integration tests.
	@echo ">>> Running MLFlow python integration tests."
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
	@ grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-24s\033[0m - %s\n", $$1, $$2}'
	@ echo

#
# Build targets
# 
PHONY: clean
clean: ## clean the go and node build artifacts
	@echo ">>> Cleaning go and node build artifacts."
	rm -Rf pkg/ui/aim/embed/repo
	rm -Rf pkg/ui/mlflow/embed/repo
	rm -Rf bin/fasttrack

PHONY: build
build: ## build the go and node components
	@echo ">>> Building go and node components."
	pkg/ui/aim/embed/build.sh
	pkg/ui/mlflow/embed/build.sh
	go build -o bin/fasttrack main.go

PHONY: run
run: build ## run the FastTrackML server
	@echo ">>> Running the FasttrackML server."
	bin/fasttrack server
