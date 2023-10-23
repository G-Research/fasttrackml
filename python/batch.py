from math import nan
from typing import List, Optional

import pandas as pd
import pyarrow as pa
from mlflow.entities import ViewType
from mlflow.exceptions import MlflowException
from mlflow.store.tracking.rest_store import RestStore
from mlflow.tracking import MlflowClient
from mlflow.tracking.fluent import (
    _get_experiment_id,
    get_experiment_by_name,
    search_experiments,
)
from mlflow.utils.rest_utils import http_request
from mlflow.utils.string_utils import is_string_type


def get_metric_histories(
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
    if not isinstance(MlflowClient()._tracking_client.store, RestStore):
        raise MlflowException("Not implemented")

    if index not in ("step", "timestamp"):
        raise ValueError(f"Unsupported index: {index}. Supported string values are 'step' or 'timestamp'")

    no_exp_ids = experiment_ids is None or len(experiment_ids) == 0
    no_exp_names = experiment_names is None or len(experiment_names) == 0
    no_run_ids = run_ids is None or len(run_ids) == 0
    no_ids_or_names = no_exp_ids and no_exp_names and no_run_ids
    if not no_exp_ids and not no_exp_names:
        raise ValueError(message="Only experiment_ids or experiment_names can be used, but not both")

    if search_all_experiments and no_ids_or_names:
        experiment_ids = [exp.experiment_id for exp in search_experiments(view_type=ViewType.ACTIVE_ONLY)]
    elif no_ids_or_names:
        experiment_ids = _get_experiment_id()
    elif not no_exp_names:
        experiments = [get_experiment_by_name(n) for n in experiment_names if n is not None]
        experiment_ids = [e.experiment_id for e in experiments if e is not None]

    if isinstance(experiment_ids, int) or is_string_type(experiment_ids):
        experiment_ids = [experiment_ids]
    if is_string_type(run_ids):
        run_ids = [run_ids]
    if is_string_type(metric_keys):
        metric_keys = [metric_keys]

    # Using an internal function as the linter doesn't like assigning a lambda, and inlining the
    # full thing is a mess
    result = http_request(
        host_creds=MlflowClient()._tracking_client.store.get_host_creds(),
        endpoint="/api/2.0/mlflow/metrics/get-histories",
        method="POST",
        json={
            "experiment_ids": experiment_ids,
            "run_ids": run_ids,
            "metric_keys": metric_keys,
            "run_view_type": ViewType.to_string(run_view_type).upper(),
            "max_results": max_results,
        },
        stream=True,
    )

    if result.status_code != 200:
        result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )

    with pa.ipc.open_stream(result.raw) as reader:
        return reader.read_pandas().set_index(["run_id", "key", index])
    