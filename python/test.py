from random import randint, random

import mlflow_wrapper
from mlflow_wrapper import log_metric, log_param

mlflow_wrapper.set_tracking_uri("http://localhost:5000/")
mlflow_wrapper.set_experiment("test-experiment2")


if __name__ == "__main__":
    # Log a parameter (key-value pair)

    log_param("param2", randint(0, 100))
    log_param("param3", ["test1", "tes2", "test3"])
    log_param("param4", {
  "brand": "Ford",
  "model": "Mustang",
  "year": 1964
})

    # Log a metric; metrics can be updated throughout the run
    log_metric("TestMetric4", 9)
    log_metric("TestMetric5", 12.2)
