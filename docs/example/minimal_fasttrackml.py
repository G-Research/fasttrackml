import time
from random import randint, random

from fasttrackml.entities.metric import Metric

import fasttrackml
from fasttrackml import FasttrackmlClient


def main():
    fasttrackml.set_tracking_uri("http://localhost:5000")
    # Creating an instance of the Fasttrackml client
    client = FasttrackmlClient()

    # Creating a new run
    experiment_id = "0"
    run = client.create_run(experiment_id)
    run_id = run.info.run_id

    # Logging a parameter
    param_key = "param1"
    param_value = randint(0, 100)
    client.log_param(run_id, param_key, param_value)

    metric_key = "foo"
    # Logging metrics with context
    client.log_metric(run_id, metric_key, random(), context={"context_key": "context_value1"})
    client.log_metric(run_id, metric_key, random() + 1, context={"context_key": "context_value2"})
    # Logging metrics without context
    client.log_metric(run_id, metric_key, random() + 2)

    # Logging a batch of metrics
    timestamp = int(time.time() * 1000)
    metrics = [
        Metric("loss", 0.2, timestamp, 1, context={"context_key": "context_value3"}),
        Metric("precision", 0.92, timestamp, 1, context={"context_key": "context_value4"}),
    ]
    client.log_batch(run_id, metrics=metrics)

    # Closing the run
    client.set_terminated(run_id)


if __name__ == "__main__":
    main()
