from typing import Dict, List, Optional, Sequence

import pandas as pd
from fasttrackml._tracking_service.client import FasttrackmlTrackingServiceClient
from fasttrackml.entities.metric import Metric
from mlflow import MlflowClient
from mlflow.entities import Param, RunTag, ViewType
from mlflow.tracking._tracking_service import utils


class FasttrackmlClient(MlflowClient):
    def __init__(self, tracking_uri: Optional[str] = None, registry_uri: Optional[str] = None):
        super().__init__(tracking_uri, registry_uri)
        final_tracking_uri = utils._resolve_tracking_uri(tracking_uri)
        self._tracking_client_mlflow = self._tracking_client
        self._tracking_client = FasttrackmlTrackingServiceClient(final_tracking_uri)

    def log_metric(
            self,
            run_id: str,
            key: str,
            value: float,
            timestamp: Optional[int] = None,
            step: Optional[int] = None,
            context: Optional[Dict[str, object]] = None,
        ) -> None:
            """
            Log a metric against the run ID.

            :param run_id: The run id to which the metric should be logged.
            :param key: Metric name (string). This string may only contain alphanumerics, underscores
                        (_), dashes (-), periods (.), spaces ( ), and slashes (/).
                        All backend stores will support keys up to length 250, but some may
                        support larger keys.
            :param value: Metric value (float). Note that some special values such
                        as +/- Infinity may be replaced by other values depending on the store. For
                        example, the SQLAlchemy store replaces +/- Inf with max / min float values.
                        All backend stores will support values up to length 5000, but some
                        may support larger values.
            :param timestamp: Time when this metric was calculated. Defaults to the current system time.
            :param step: Integer training step (iteration) at which was the metric calculated.
                        Defaults to 0.
            :param context: Additional context about the metric. This is a dictionary of
                        key-value pairs whose values are strings. 

            .. code-block:: python
                :caption: Example

                from fasttrackml import FasttrackmlClient


                def print_run_info(r):
                    print(f"run_id: {r.info.run_id}")
                    print(f"metrics: {r.data.metrics}")
                    print(f"status: {r.info.status}")


                # Create a run under the default experiment (whose id is '0').
                # Since these are low-level CRUD operations, this method will create a run.
                # To end the run, you'll have to explicitly end it.
                client = FasttrackmlClient()
                experiment_id = "0"
                run = client.create_run(experiment_id)
                print_run_info(run)
                print("--")

                # Log the metric. Unlike mlflow.log_metric this method
                # does not start a run if one does not exist. It will log
                # the metric for the run id in the backend store.
                client.log_metric(run.info.run_id, "m", 1.5, context={"context_key": "context_value"})
                client.set_terminated(run.info.run_id)
                run = client.get_run(run.info.run_id)
                print_run_info(run)

            .. code-block:: text
                :caption: Output

                run_id: 95e79843cb2c463187043d9065185e24
                metrics: {}
                status: RUNNING
                --
                run_id: 95e79843cb2c463187043d9065185e24
                metrics: {'m': 1.5}
                status: FINISHED
            """
            self._tracking_client.log_metric(run_id, key, value, timestamp, step, context)
    
    def log_batch(
            self,
            run_id: str,
            metrics: Sequence[Metric] = (),
            params: Sequence[Param] = (),
            tags: Sequence[RunTag] = (),
        ) -> None:
            """
            Log multiple metrics, params, and/or tags.

            :param run_id: String ID of the run
            :param metrics: If provided, List of Metric(key, value, timestamp) instances.
            :param params: If provided, List of Param(key, value) instances.
            :param tags: If provided, List of RunTag(key, value) instances.


            Raises an MlflowException if any errors occur.
            :return: None

            .. code-block:: python
                :caption: Example

                import time

                from fasttrackml import FasttrackmlClient
                from fasttrackml.entities import Metric, Param, RunTag


                def print_run_info(r):
                    print(f"run_id: {r.info.run_id}")
                    print(f"params: {r.data.params}")
                    print(f"metrics: {r.data.metrics}")
                    print(f"tags: {r.data.tags}")
                    print(f"status: {r.info.status}")


                # Create MLflow entities and a run under the default experiment (whose id is '0').
                timestamp = int(time.time() * 1000)
                metrics = [Metric("m", 1.5, timestamp, 1, {"context_key": "context_value"})]
                params = [Param("p", "p")]
                tags = [RunTag("t", "t")]
                experiment_id = "0"
                client = FasttrackmlClient()
                run = client.create_run(experiment_id)

                # Log entities, terminate the run, and fetch run status
                client.log_batch(run.info.run_id, metrics=metrics, params=params, tags=tags)
                client.set_terminated(run.info.run_id)
                run = client.get_run(run.info.run_id)
                print_run_info(run)

            .. code-block:: text
                :caption: Output

                run_id: ef0247fa3205410595acc0f30f620871
                params: {'p': 'p'}
                metrics: {'m': 1.5}
                tags: {'t': 't'}
                status: FINISHED
            """
            self._tracking_client_mlflow.log_batch(run_id, params=params, tags=tags)
            self._tracking_client.log_batch(run_id, metrics, params, tags)
