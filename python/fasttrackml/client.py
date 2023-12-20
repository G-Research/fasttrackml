from typing import Dict, List, Optional, Sequence

import pandas as pd
from fasttrackml._tracking_service.client import FasttrackmlTrackingServiceClient
from fasttrackml.entities.metric import Metric
from mlflow import MlflowClient
from mlflow.entities import ViewType
from mlflow.tracking._tracking_service import utils


class FasttrackmlClient(MlflowClient):
    def __init__(self, tracking_uri: Optional[str] = None, registry_uri: Optional[str] = None):
        super().__init__(tracking_uri, registry_uri)
        final_tracking_uri = utils._resolve_tracking_uri(tracking_uri)
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
            self._tracking_client.log_metric(run_id, key, value, timestamp, step, context)
    
    def log_batch(
            self,
            run_id: str,
            metrics: Sequence[Metric] = (),
        ) -> None:
            self._tracking_client.log_batch(run_id, metrics)

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
        :return: ``pandas.DataFrame`` of metric timestamps and values, indexed on run ID, metric key,
                and step. If index is ``timestamp``, the columns will be metric steps and values, and
                the index will be run ID, metric key, and timestamp.

        """
        return self._tracking_client.get_metric_histories(experiment_ids, run_ids, metric_keys, index, run_view_type, max_results, search_all_experiments, experiment_names)