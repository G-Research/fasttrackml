from typing import Dict, List, Optional, Sequence

import pandas as pd
from mlflow import MlflowClient
from mlflow.entities import Param, RunTag, ViewType
from mlflow.tracking._tracking_service import utils

from ._tracking_service.client import FasttrackmlTrackingServiceClient
from .entities.metric import Metric


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

    def get_metric_history(self, run_id: str, key: str) -> List[Metric]:
        """
        Return a list of metric objects corresponding to all values logged for a given metric.

        :param run_id: Unique identifier for run
        :param key: Metric name within the run

        :return: A list of :py:class:`mlflow.entities.Metric` entities if logged, else empty list

        .. code-block:: python
            :caption: Example

            from fasttrackml import FasttrackmlClient


            def print_metric_info(history):
                for m in history:
                    print(f"name: {m.key}")
                    print(f"value: {m.value}")
                    print(f"step: {m.step}")
                    print(f"timestamp: {m.timestamp}")
                    print(f"context: {m.context}")
                    print("--")


            # Create a run under the default experiment (whose id is "0"). Since this is low-level
            # CRUD operation, the method will create a run. To end the run, you'll have
            # to explicitly end it.
            client = FasttrackmlClient()
            experiment_id = "0"
            run = client.create_run(experiment_id)
            print(f"run_id: {run.info.run_id}")
            print("--")

            # Log couple of metrics, update their initial value, and fetch each
            # logged metrics' history.
            for k, v in [("m1", 1.5), ("m2", 2.5)]:
                client.log_metric(run.info.run_id, k, v, step=0, context={"context_key": "context_value"})
                client.log_metric(run.info.run_id, k, v + 1, step=1, context={"context_key1": "context_value1"})
                print_metric_info(client.get_metric_history(run.info.run_id, k))
            client.set_terminated(run.info.run_id)

        .. code-block:: text
            :caption: Output

            run_id: c360d15714994c388b504fe09ea3c234
            --
            name: m1
            value: 1.5
            step: 0
            timestamp: 1603423788607
            context: {'context_key': 'context_value'}
            --
            name: m1
            value: 2.5
            step: 1
            timestamp: 1603423788608
            context: {'context_key1': 'context_value1'}
            --
            name: m2
            value: 2.5
            step: 0
            timestamp: 1603423788609
            context: {'context_key': 'context_value'}
            --
            name: m2
            value: 3.5
            step: 1
            timestamp: 1603423788610
            context: {'context_key1': 'context_value1'}
            --
        """
        return self._tracking_client.get_metric_history(run_id, key)

    def get_metric_histories(
        self,
        experiment_ids: Optional[List[str]] = None,
        run_ids: Optional[List[str]] = None,
        metric_keys: Optional[List[str]] = None,
        index: str = "step",
        run_view_type: int = ViewType.ACTIVE_ONLY,
        max_results: int = 10000000,
        search_all_experiments: bool = False,
        experiment_names: Optional[List[str]] = None,
        context: Optional[Dict[str, object]] = None,
    ) -> pd.DataFrame:
        """
        Get metric histories of Runs that fit the specified criteria.

        :param experiment_ids: List of experiment IDs. Search can work with experiment IDs or
                            experiment names, but not both in the same call. Values other than
                            ``None`` or ``[]`` will result in error if ``experiment_names`` is
                            also not ``None`` or ``[]``. ``None`` will default to the active
                            experiment if ``experiment_names`` is ``None`` or ``[]``.
        :param run_ids: List of run IDs to get metric histories for. Cannot be specified at the same
                        time as experiment_names or experiment_ids.
        :param metric_keys: List of metric keys to get the metric histories for.
        :param run_view_type: One of enum values ``ACTIVE_ONLY``, ``DELETED_ONLY``, or ``ALL`` runs
                            defined in :py:class:`mlflow.entities.ViewType`.
        :param max_results: The maximum number of metric values to put in the dataframe. Default is
                            10,000,000 to avoid causing out-of-memory issues on the user's machine.
        :param search_all_experiments: Boolean specifying whether all experiments should be searched.
                                    Only honored if ``experiment_ids`` is ``[]`` or ``None``.
        :param experiment_names: List of experiment names. Search can work with experiment IDs or
                                experiment names, but not both in the same call. Values other
                                than ``None`` or ``[]`` will result in error if ``experiment_ids``
                                is also not ``None`` or ``[]``. ``None`` will default to the active
                                experiment if ``experiment_ids`` is ``None`` or ``[]``.
        :param context: Dictionary of json paths (keys) and values which must be found in the
                        metric context recorded when logged.
        :return: ``pandas.DataFrame`` of metric timestamps and values, indexed on run ID, metric key,
                and step. If index is ``timestamp``, the columns will be metric steps and values, and
                the index will be run ID, metric key, and timestamp.

        """
        return self._tracking_client.get_metric_histories(
            experiment_ids,
            run_ids,
            metric_keys,
            index,
            run_view_type,
            max_results,
            search_all_experiments,
            experiment_names,
            context,
        )
