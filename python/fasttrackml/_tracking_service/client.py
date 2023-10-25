from fasttrackml.entities.metric_context import MetricContext
from fasttrackml.entities.metric_with_context import MetricWithContext
from fasttrackml.store.contextsupport_store import ContextsupportStore
from mlflow.tracking._tracking_service.client import TrackingServiceClient
from mlflow.tracking.metric_value_conversion_utils import (
    convert_metric_value_to_float_if_possible,
)
from mlflow.utils.rest_utils import MlflowHostCreds
from mlflow.utils.time import get_current_time_millis


class TrackingServiceClientExtend(TrackingServiceClient):

    def __init__(self, tracking_uri):
        super().__init__(tracking_uri)

    def log_metric_with_context(self, run_id, key, value, timestamp=None, step=None, context=None):
        timestamp = timestamp if timestamp is not None else get_current_time_millis()
        step = step if step is not None else 0
        context = context if context else {}
        metric_value = convert_metric_value_to_float_if_possible(value)
        contextList=[MetricContext(key, value) for key, value in context.items()]
        metric = MetricWithContext(key, metric_value, timestamp, step, contextList)

        store = ContextsupportStore(lambda: MlflowHostCreds(self.tracking_uri))
        store.log_metric_with_context(run_id, metric)