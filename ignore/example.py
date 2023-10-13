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
