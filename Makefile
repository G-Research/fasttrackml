## Fasttrack Makefile
## For best results, run these targets inside the devcontainer

PHONY: build
build:
	## build the javascript and go binary
	pkg/ui/aim/embed/build.sh
	pkg/ui/mlflow/embed/build.sh
	go build -o bin/fasttrack main.go

PHONY: run
run: build
	## Run the fasttrack server
	bin/fasttrack server

PHONY: test
test: build
	## run the integration tests from mlflow
	tests/mlflow/test.sh
