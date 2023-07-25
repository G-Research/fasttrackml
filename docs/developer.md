# Developer Guide

## Dev Container

Install the [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension for VS Code. The extension will automatically detect the `.devcontainer` folder and prompt you to open the project in a container.

## Classic

### Install Dependencies

FastTrackML requires the following dependencies to be installed on your system:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)

### Build FastTrackML

FastTrackML can be built using the following command:

```bash
make build
```