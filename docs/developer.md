# Developer Guide

## Dev Container

Install the [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension for VS Code. The extension will automatically detect the `.devcontainer` folder and prompt you to open the project in a container.

## Classic

### Install Dependencies

FastTrack requires the following dependencies to be installed on your system:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)

### Build FastTrack

FastTrack can be built using the following command:

```bash
# Install json parser
sudo apt-get install jq

# Get the build tags and version from the settings.json file
tags="$(jq -r '."go.buildTags"' .vscode/settings.json)"
version=$(git describe --tags --dirty | sed 's/^v//')

docker build --build-arg tags="$tags" --build-arg version="$version" -t fasttrack .
```