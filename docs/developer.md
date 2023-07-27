# Developer Guide

## Dev Container

Install the [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension for VS Code. The extension will automatically detect the `.devcontainer` folder and prompt you to open the project in a container.

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