# Developer Guide

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