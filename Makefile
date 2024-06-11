## FastTrackML
## For best results, run these make targets inside the devcontainer

#
# Project-specific variables
#
# App name.
APP=fml
ifeq ($(shell go env GOOS),windows)
  APP:=$(APP).exe
endif
# Version.
# Use git describe to get the version.
# If the git describe fails, fallback to a version based on the git commit.
VERSION?=$(shell git describe --tags --dirty --match='v*' 2> /dev/null | sed 's/^v//')
ifeq ($(VERSION),)
  VERSION=0.0.0-g$(shell git describe --always --dirty 2> /dev/null)
endif
# Go ldflags.
# Strip debug symbols and disable DWARF generation.
# Build static binaries on Linux.
GO_LDFLAGS=-s -w -X github.com/G-Research/fasttrackml/pkg/version.Version=$(VERSION)
ifeq ($(shell go env GOOS),linux)
  GO_LDFLAGS+=-linkmode external -extldflags -static
endif
# Go build tags.
GO_BUILDTAGS=$(shell cat .go-build-tags 2> /dev/null)
# Archive information.
# Use zip on Windows, tar.gz on Linux and macOS.
# Use GNU tar on macOS if available, to avoid issues with BSD tar.
ifeq ($(shell go env GOOS),windows)
  ARCHIVE_EXT=zip
  ARCHIVE_CMD=zip -r
else
  ARCHIVE_EXT=tar.gz
  ARCHIVE_CMD=tar -czf
  ifeq ($(shell which gtar >/dev/null 2>/dev/null; echo $$?),0)
    ARCHIVE_CMD:=g$(ARCHIVE_CMD)
  endif
endif
ARCHIVE_NAME=dist/fasttrackml_$(shell go env GOOS | sed s/darwin/macos/)_$(shell go env GOARCH | sed s/amd64/x86_64/).$(ARCHIVE_EXT)
ARCHIVE_FILES=$(APP) LICENSE README.md
# Docker compose file.
COMPOSE_FILE=tests/integration/docker-compose.yml
# Docker compose project name.
COMPOSE_PROJECT_NAME=$(APP)-integration-tests

AIM_BUILD_LOCATION=$(HOME)/fasttrackml-ui-aim
MLFLOW_BUILD_LOCATION=$(HOME)/fasttrackml-ui-mlflow

#
# Default target (help).
#
.PHONY: help
help: ## display this help.
	@echo "Please use \`make <target>' where <target> is one of:"
	@echo
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "    \033[36m%-30s\033[0m - %s\n", $$1, $$2}'
	@echo

#
# Tools targets.
# This needs to be kept in sync with .devcontainer/Dockerfile.
#
.PHONY: install-tools
install-tools: ## install tools.
	@echo '>>> Installing tools.'
	@go install github.com/vektra/mockery/v2@v2.34.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
	@go install golang.org/x/tools/cmd/goimports@v0.13.0
	@go install mvdan.cc/gofumpt@v0.5.0

#
# Linter targets.
#
.PHONY: lint
lint: go-lint python-lint ## run set of linters over the code.

#
# Go targets.
#
.PHONY: go-get
go-get: ## get go modules.
	@echo '>>> Getting go modules.'
	@go mod download

.PHONY: go-build
go-build: ## build app binary.
	@echo '>>> Building go binary.'
	@CGO_ENABLED=1 go build -ldflags="$(GO_LDFLAGS)" -tags="$(GO_BUILDTAGS)" -o $(APP) ./main.go

.PHONY: go-format
go-format: ## format go code.
	@echo '>>> Formatting go code.'
	@gofumpt -w .
	@goimports -w -local github.com/G-Research/fasttrackml $(shell find . -type f -name '*.go' -not -name 'mock_*.go')

.PHONY: go-lint
go-lint: ## run go linters.
	@echo '>>> Running go linters.'
	@golangci-lint run -v --build-tags $(GO_BUILDTAGS)

.PHONY: go-dist
go-dist: go-build ## archive app binary.
	@echo '>>> Archiving go binary.'
	@dir=$$(dirname $(ARCHIVE_NAME)); if [ ! -d $$dir ]; then mkdir -p $$dir; fi
	@if [ -f $(ARCHIVE_NAME) ]; then rm -f $(ARCHIVE_NAME); fi
	@$(ARCHIVE_CMD) $(ARCHIVE_NAME) $(ARCHIVE_FILES)

#
# Python targets.
#
.PHONY: python-env
python-env: ## create python virtual environment.
	@echo '>>> Creating python virtual environment.'
	@pipenv sync

.PHONY: python-dist
python-dist: go-build python-env ## build python wheels.
	@echo '>>> Building Python Wheels.'
	@VERSION=$(VERSION) pipenv run python3 -m pip wheel ./python --wheel-dir=wheelhouse --no-deps

.PHONY: python-format
python-format: python-env ## format python code.
	@echo '>>> Formatting python code.'
	@pipenv run black --line-length 120 .
	@pipenv run isort --profile black .

.PHONY: python-lint
python-lint: python-env ## check python code formatting.
	@echo '>>> Checking python code formatting.'
	@pipenv run black --check --line-length 120 .
	@pipenv run isort --check-only --profile black .

#
# Tests targets.
#
.PHONY: test
test: test-go-unit container-test test-python-integration ## run all the tests.

.PHONY: test-go-unit
test-go-unit: ## run go unit tests.
	@echo ">>> Running unit tests."
	@go test -tags="$(GO_BUILDTAGS)" ./pkg/...

.PHONY: test-go-integration
test-go-integration: ## run go integration tests.
	@echo ">>> Running integration tests."
	@go test -tags="$(GO_BUILDTAGS)" ./tests/integration/golang/...

.PHONY: test-go-compatibility
test-go-compatibility: ## run go compatibility tests.
	@echo ">>> Running compatibility tests."
	@go test -tags="$(GO_BUILDTAGS),compatibility" ./tests/integration/golang/compatibility

.PHONY: test-python-integration
test-python-integration: ## run all the python integration tests.
	@echo ">>> Running all python integration tests."
	@go run tests/integration/python/main.go

.PHONY: test-python-integration-mlflow
test-python-integration-mlflow: ## run the MLFlow python integration tests.
	@echo ">>> Running MLFlow python integration tests."
	@go run tests/integration/python/main.go -targets mlflow

.PHONY: test-python-integration-aim
test-python-integration-aim: ## run the Aim python integration tests.
	@echo ">>> Running Aim python integration tests."
	@go run tests/integration/python/main.go -targets aim

.PHONY: test-python-integration-client
test-python-integration-client: ## run the FML Client python integration tests.
	@echo ">>> Running FasttrackmlClient python integration tests."
	@go run tests/integration/python/main.go -targets fml_client

#
# Container test targets.
#
.PHONY: container-test
container-test: ## run integration tests in container.
	@echo ">>> Running integration tests in container."
	@COMPOSE_FILE=$(COMPOSE_FILE) COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) \
		docker compose run -e FML_SLOW_TESTS_ENABLED integration-tests

.PHONY: container-compatibility-test
container-compatibility-test: ## run compatibility tests in container.
	@echo ">>> Running compatibility tests in container."
	@COMPOSE_FILE=$(COMPOSE_FILE) COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) \
		docker compose run -e MLFLOW_VERSION -e DATABASE_URI mlflow-setup
	@COMPOSE_FILE=$(COMPOSE_FILE) COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) \
		docker compose run -e MLFLOW_VERSION -e DATABASE_URI compatibility-tests

.PHONY: container-clean
container-clean: ## clean containers.
	@echo ">>> Cleaning containers."
	@COMPOSE_FILE=$(COMPOSE_FILE) COMPOSE_PROJECT_NAME=$(COMPOSE_PROJECT_NAME) \
		docker compose down -v --remove-orphans

#
# Mockery targets.
#
.PHONY: mocks-clean
mocks-clean: ## cleans mocks.
	@echo ">>> Cleaning mocks."
	@find ./pkg -name 'mock_*.go' -type f -delete

.PHONY: mocks-generate
mocks-generate: mocks-clean ## generate mock based on all project interfaces.
	@echo ">>> Generating mocks."
	@mockery

#
# Docker targets (Only available on Linux).
#
ifeq ($(shell go env GOOS),linux)
# Load into the Docker daemon by default.
DOCKER_OUTPUT?=type=docker
ifneq ($(origin DOCKER_METADATA), undefined)
  # If DOCKER_METADATA is defined, use it to set the tags and labels.
  # DOCKER_METADATA should be a JSON object with the following structure:
  # {
  #   "tags": ["image:tag1", "image:tag2"],
  #   "labels": {
  #     "label1": "value1",
  #     "label2": "value2"
  #   }
  # }
  DOCKER_TAGS=$(shell echo $$DOCKER_METADATA | jq -r '.tags | map("--tag \(.)") | join(" ")')
  DOCKER_LABELS=$(shell echo $$DOCKER_METADATA | jq -r '.labels | to_entries | map("--label \(.key)=\"\(.value)\"") | join(" ")')
else
  # Otherwise, use DOCKER_TAGS if defined, otherwise use the default.
  # DOCKER_TAGS should be a space-separated list of tags.
  # e.g. DOCKER_TAGS="image:tag1 image:tag2"
  # We do not set DOCKER_LABELS because of the way make handles spaces
  # in variable values. Use DOCKER_METADATA if you need to set labels.
  DOCKER_TAGS?=fasttrackml:$(VERSION) fasttrackml:latest
  DOCKER_TAGS:=$(addprefix --tag ,$(DOCKER_TAGS))
endif
.PHONY: docker-dist
docker-dist: go-build ## build docker image.
	@echo ">>> Building Docker image."
	@docker buildx build --provenance false --sbom false --platform linux/$(shell go env GOARCH) --output $(DOCKER_OUTPUT) $(DOCKER_TAGS) $(DOCKER_LABELS) .
endif

#
# Build targets.
#
.PHONY: clean
clean: ## clean build artifacts.
	@echo ">>> Cleaning build artifacts."
	@rm -rf $(APP) dist wheelhouse

.PHONY: build
build: go-build ## build the app.

.PHONY: dist
dist: go-dist python-dist ## build the software archives.
ifeq ($(shell go env GOOS),linux)
dist: docker-dist
endif

.PHONY: format
format: go-format python-format ## format the code.

.PHONY: run
run: build ## run the FastTrackML server.
	@echo ">>> Running the FasttrackML server."
	@./$(APP) server

.PHONY: migrations-create
migrations-create: ## generate a new database migration.
	@echo ">>> Running FastTrackML migrations create."
	@go run main.go migrations create

.PHONY: migrations-rebuild
migrations-rebuild: ## rebuild the migrations script to detect new migrations.
	@echo ">>> Running FastTrackML migrations rebuild."
	@go run main.go migrations rebuild

.PHONY: ui-aim-sync
ui-aim-sync: ## copy Aim UI files to docker volume.
	@echo ">>> Syncing the Aim UI."
	@rsync -rvu --exclude node_modules --exclude .git ui/fasttrackml-ui-aim/ $(AIM_BUILD_LOCATION)

.PHONY: ui-aim-start
ui-aim-start: ui-aim-sync ## start the Aim UI for development.  
	@echo ">>> Starting the Aim UI."
	@cd $(AIM_BUILD_LOCATION)/src && npm ci --legacy-peer-deps && npm start

.PHONY: ui-mlflow-sync
ui-mlflow-sync: ## copy MLflow UI files to docker volume.
	@echo ">>> Syncing the MLflow UI."
	@rsync -rvu --exclude node_modules --exclude .git ui/fasttrackml-ui-mlflow/ $(MLFLOW_BUILD_LOCATION)

.PHONY: ui-mlflow-start
ui-mlflow-start: ui-mlflow-sync ## start the MLflow UI for development.
	@echo ">>> Starting the MLflow UI."
	@cd $(MLFLOW_BUILD_LOCATION)/src && yarn install --immutable && yarn start
