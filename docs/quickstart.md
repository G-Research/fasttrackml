# FastTrack Quickstart

## Install Dependencies

FastTrack requires the following dependencies to be installed on your system:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)
- [Python 3](https://www.python.org/downloads/)
- [MLFlow](https://mlflow.org/docs/latest/index.html)

## Build FastTrack

FastTrack can be built using the following command:

```bash
# Install json parser
sudo apt-get install jq

# Get the build tags and version from the settings.json file
tags="$(jq -r '."go.buildTags"' .vscode/settings.json)"
version=$(git describe --tags | sed 's/^v//')

docker build --build-arg tags="$tags" --build-arg version="$version" -t fasttrack .
```

## Run FastTrack

FastTrack can be run using the following command:

```bash
docker run --rm -p 5000:5000 -ti fasttrack
```

Verify that you can see the UI by navigating to http://localhost:5000/.

![FastTrack UI](https://files.mcaq.me/57b05.jpg)


## Run a quick test script

```bash
python3 ./docs/dev/minimal.py
```

After running this script, you should see the following output from http://localhost:5000/aim/:

![FastTrack UI](https://files.mcaq.me/43x5j.jpg)

From here you can check out the metrics and run information to see more details about the run.

## Testing a Random Forrest Model

```bash
# Install mflow and poetry
cd docs/dev
poetry install
# MLFlow will not be installed by poetry, so we need to install it manually
poetry run pip install mlflow boto3

# Run the script
poetry run python3 random_forrest.py
```

