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

## Running a real ML experiment

An example is included that uses Stable Baseline 3 - a popular reinforcement learning library. It is running the CartPole Environment with a PPO agent.

You can read about the how it works [here](https://gsurma.medium.com/cartpole-introduction-to-reinforcement-learning-ed0eb5b58288#c876).

Run the example with:

```bash
poetry run python3 stable_baseline.py
```

And then you will see the experiment `cartpole-ppo-v1` and be able to view the metrics.

You can also see the videos of the model in `docs/example/vidcap` as it trains, as well as the final model weights in `ckpt`
