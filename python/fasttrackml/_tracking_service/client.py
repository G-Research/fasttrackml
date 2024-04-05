from functools import partial
from typing import Dict, Optional, Sequence

from mlflow.entities import Param, RunTag
from mlflow.store.tracking import GET_METRIC_HISTORY_MAX_RESULTS
from mlflow.tracking._tracking_service.client import TrackingServiceClient
from mlflow.tracking.metric_value_conversion_utils import (
    convert_metric_value_to_float_if_possible,
)
from mlflow.utils import chunk_list

try:
    from mlflow.utils.credentials import get_default_host_creds
except ImportError:
    from mlflow.tracking._tracking_service.utils import _get_default_host_creds as get_default_host_creds

from mlflow.utils.time import get_current_time_millis
from mlflow.utils.validation import MAX_METRICS_PER_BATCH

from ..entities.metric import Metric
from ..store.custom_rest_store import CustomRestStore


class FasttrackmlTrackingServiceClient(TrackingServiceClient):

    def __init__(self, tracking_uri):
        super().__init__(tracking_uri)
        self.custom_store = CustomRestStore(partial(get_default_host_creds, self.tracking_uri))

    def log_metric(
        self,
        run_id: str,
        key: str,
        value: float,
        timestamp: Optional[int] = None,
        step: Optional[int] = None,
        context: Optional[dict] = None,
    ):
        timestamp = timestamp if timestamp is not None else get_current_time_millis()
        step = step if step is not None else 0
        context = context if context else {}
        metric_value = convert_metric_value_to_float_if_possible(value)
        metric = Metric(key, metric_value, timestamp, step, context)
        self.custom_store.log_metric(run_id, metric)

    def log_batch(
        self, run_id: str, metrics: Sequence[Metric] = (), params: Sequence[Param] = (), tags: Sequence[RunTag] = ()
    ):
        for metrics_batch in chunk_list(metrics, chunk_size=MAX_METRICS_PER_BATCH):
            self.custom_store.log_batch(run_id=run_id, metrics=metrics_batch)

    def get_metric_history(self, run_id, key):
        # NB: Paginated query support is currently only available for the RestStore backend.
        # FileStore and SQLAlchemy store do not provide support for paginated queries and will
        # raise an MlflowException if the `page_token` argument is not None when calling this
        # API for a continuation query.
        history = self.custom_store.get_metric_history(
            run_id=run_id,
            metric_key=key,
            max_results=GET_METRIC_HISTORY_MAX_RESULTS,
            page_token=None,
        )
        token = history.token
        # Continue issuing queries to the backend store to retrieve all pages of
        # metric history.
        while token is not None:
            paged_history = self.store.get_metric_history(
                run_id=run_id,
                metric_key=key,
                max_results=GET_METRIC_HISTORY_MAX_RESULTS,
                page_token=token,
            )
            history.extend(paged_history)
            token = paged_history.token
        return history

    def get_metric_histories(
        self,
        experiment_ids: Optional[Sequence[str]] = None,
        run_ids: Optional[Sequence[str]] = None,
        metric_keys: Optional[Sequence[str]] = None,
        index: str = "step",
        run_view_type: int = None,
        max_results: int = 10000000,
        search_all_experiments: bool = False,
        experiment_names: Optional[Sequence[str]] = None,
        context: Optional[Dict[str, object]] = None,
    ):
        return self.custom_store.get_metric_histories(
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
