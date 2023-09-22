# FastTrackML Quickstart

## Install FastTrackML

### With `pip`

```bash
pip install fasttrackml
```

### With a script

#### On Linux and macOS

```bash
curl -fsSL https://fasttrackml.io/install.sh | sh
```

#### On Windows

```bash
iwr -useb https://fasttrackml.io/install.ps1 | iex
```

### Manually

Download the executable for your platform from the [latest release](https://github.com/G-Research/fasttrackml/releases/latest) assets.
Extract it and then validate your installation with the following command:

```bash
./fml --version
```

## Run FastTrackML

### Natively

```bash
fml server
```

### Via Docker

You can also run FastTrackML in a container via [Docker](https://docs.docker.com/get-docker/):

```bash
docker run --rm -p 5000:5000 -ti gresearch/fasttrackml
```

### Verification

Verify that you can see the UI by navigating to http://localhost:5000/.

![FastTrackML UI](images/main_ui.png)

## Run a quick test script

To run the test scripts, you need a working [Python](https://www.python.org/downloads/) installation and the [Poetry](https://python-poetry.org/docs/#installation) package manager.

```bash
# Install mflow and poetry
cd docs/example
poetry install
poetry run python3 minimal.py
```

After running this script, you should see the following output from http://localhost:5000/aim/:

![FastTrackML UI](images/runs_ui.png)

From here you can check out the metrics and run information to see more details about the run.

## Testing a Random Forest Model

**Note that since artifacts are not yet supported, most of the autolog features will not work.**

### Get the required data

From Kaggle, download https://www.kaggle.com/datasets/kyr7plus/emg-4?resource=download

Extract the zip file and move the files to `docs/example/data`.

### Run the script

```bash
poetry run python3 random_forest.py
```

