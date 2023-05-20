## Run FastTrackML

> <small><strong>Note</strong>: This step requires <a href="https://docs.docker.com/get-docker/">Docker</a> to be
> running.</small>

FastTrackML can be run using the following command:

```bash
docker run --rm -p 5000:5000 -ti gresearch/fasttrackml
```

Verify that you can see the UI by navigating to [http://localhost:5000/](http://localhost:5000/).

![FastTrackML UI](images/main_ui.jpg)

## Track your first experiment

> <small><strong>Note</strong>: This step requires <a href="https://www.python.org/downloads/">Python 3</a> to be
> installed.</small>

Install the ML Flow Python package:

```bash
pip install mlflow
```

Then, run the following Python script to log a parameter and metric to FastTrackML:

```python
import mlflow
import random

# Set the tracking URI to the FastTrackML server
mlflow.set_tracking_uri("http://localhost:5000")
# Set the experiment name
mlflow.set_experiment("my-experiment")

# Log a parameter
mlflow.log_param("param1", random.randint(0, 100))

# Log a metric
mlflow.log_metric("foo", random.random())
# metrics can be updated throughout the run
mlflow.log_metric("foo", random.random() + 1)
mlflow.log_metric("foo", random.random() + 2)
```

After running this script, you should see the following output
from [http://localhost:5000/aim/](http://localhost:5000/aim/):

![FastTrackML UI](images/runs_ui.jpg)

From here you can check out the metrics and run information to see more details about the run.
