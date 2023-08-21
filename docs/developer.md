# Developer Guide

## Dev Container

Install the [Dev
Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
extension for VS Code. The extension will automatically detect the
`.devcontainer` folder and prompt you to open the project in a
container.

## Classic

### Install Dependencies

FastTrackML requires the following dependencies to be installed on your system:

- [Go SDK](https://go.dev/dl/)
- A working C compiler for your platform
  - macOS: `xcode-select --install`
  - Debian/Ubuntu: `sudo apt install build-essential`
  - Windows: Install [MSYS2](https://www.msys2.org)

### Build FastTrackML

FastTrackML can be built using the following command:

```bash
make build
```

## Working with the UIs

FasttrackML incorporates the existing Aim and MLFlow web UIs, albeit
with a few modifications. This is accomplished at build time by
importing the `fasttrackml-ui-aim` and `fasttrackml-ui-mlflow`
modules. These repos contain patched and compiled ui assets of the upstream
repos. To make a UI change, a PR is merged to the appropriate
release branch and new tag is pushed. At that point, the `fasttrackml`
reference can be updated (in go.mod) to the new tag.

For UI development, you'll need a tighter change/view loop,
so we recommend the following approach.

Prerequisites:
1. go 1.20 or higher
2. dagger package (`go get dagger.io/dagger`)
3. docker

Steps:
1. Clone the UI repo as a sibling to your `fasttrackml` working copy.
2. Change the UI working copy to a release branch, eg `release/v3.1.6`
3. In the `fasttrackml` folder, use `go work` to map your UI working copy
as a go module replacement:
```bash
go work init
go work use .
go work use ../fasttrackml-ui-mlflow
```
4. Mount both directories under `/workspaces` in the devcontainer. Examples
of this are commented in the `.devcontainer/docker-compose.yaml` file.
5. In your host system, compile the UI project:
```bash
cd ../fasttrackml-ui-mlflow
cd builder
go run main.go
```
6. On success, you can now start the `fasttrackml` server using the assets of 
your UI working copy.
```bash
cd ../fasttrackml
go run main.go server --listen-address ":5000"
```
