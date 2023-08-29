# FastTrackML Quickstart

## Install Dependencies

FastTrackML requires the following _optional_ dependencies to be installed on your system:

- [Docker](https://docs.docker.com/get-docker/) — enables you to download and run FastTrackML with just one command
- [Python 3](https://www.python.org/downloads/) — enables you to run the test scripts

## Run FastTrackML

### With Docker

FastTrackML can be run using the following command:

```bash
docker run --rm -p 5000:5000 -ti gresearch/fasttrackml
```

### With a native executable

Download the executable for your platform from the [latest release](https://github.com/G-Research/fasttrackml/releases/latest) assets.
Extract it and then run FastTrackML with the following command:

```bash
./fml
```

### Verification

Verify that you can see the UI by navigating to http://localhost:5000/.

![FastTrackML UI](images/main_ui.png)

## Run a quick test script

```bash
# Install mflow and poetry
cd docs/example
poetry install
# MLFlow will not be installed by poetry, so we need to install it manually
poetry run pip install mlflow boto3

python3 minimal.py
```

After running this script, you should see the following output from http://localhost:5000/aim/:

![FastTrackML UI](images/runs_ui.png)

From here you can check out the metrics and run information to see more details about the run.

## Testing a Random Forest Model

**Note that since artifacts are not yet supported, most of the autolog features will not work.**

### Get the required data

From Kaggle, download https://www.kaggle.com/datasets/kyr7plus/emg-4?resource=download

Extract the zip file and move the files to `docs/example/data`.

```bash
# Install mflow and poetry
cd docs/example
poetry install
# MLFlow will not be installed by poetry, so we need to install it manually
poetry run pip install mlflow boto3

# Run the script
poetry run python3 random_forest.py
```

