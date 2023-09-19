[![FastTrackML banner](https://fasttrackml.io/images/github-banner.svg)](https://fasttrackml.io/)

# _FastTrackML_

An experiment tracking server focused on speed and scalability, fully compatible with MLFlow.

### Quickstart

#### Run the tracking server

> [!NOTE]
> For the full guide, see [docs/quickstart.md](docs/quickstart.md).

FastTrackML can be installed and run with `pip`:

```bash
pip install fasttrackml
fml server
```

Alternatively, you can run it within a container with
[Docker](https://docs.docker.com/get-docker/):

```bash
docker run --rm -p 5000:5000 -ti gresearch/fasttrackml
```

Verify that you can see the UI by navigating to http://localhost:5000/.

![FastTrackML UI](https://raw.githubusercontent.com/G-Research/fasttrackml/main/docs/images/main_ui.png)

For more info, `--help` is your friend!

#### Track your experiments

Install the MLFlow Python package:

```bash
pip install mlflow-skinny
```

Here is an elementary example Python script:

```python
import mlflow
import random

# Set the tracking URI to the FastTrackML server
mlflow.set_tracking_uri("http://localhost:5000")
# Set the experiment name
mlflow.set_experiment("my-first-experiment")

# Start a new run
with mlflow.start_run():
    # Log a parameter
    mlflow.log_param("param1", random.randint(0, 100))

    # Log a metric
    mlflow.log_metric("foo", random.random())
    # metrics can be updated throughout the run
    mlflow.log_metric("foo", random.random() + 1)
    mlflow.log_metric("foo", random.random() + 2)
```




### Developer

Using the project's devcontainer is recommended for development. VSCode should detect
the .devcontainer folder and offer to restart the IDE in that context. For other users,
the underlying docker container can be used. The Makefile offers some basic targets.

```
cd .devcontainer
docker-compose up -d
docker-compose exec -w /workspaces/fasttrackml app bash

root ➜ /workspaces/fastrackml $ make build
root ➜ /workspaces/fastrackml $ make run
root ➜ /workspaces/fastrackml $ make test
root ➜ /workspaces/fastrackml $ emacs .
```

Note that on MacOS, port 5000 is already occupied, so some [adjustments](https://apple.stackexchange.com/a/431164) are necessary.

### License

Copyright 2022-2023 G-Research

Copyright 2019-2022 Aimhub, Inc.

Copyright 2018 Databricks, Inc.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use these files except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
