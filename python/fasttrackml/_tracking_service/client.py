from fasttrackml.store.contextsupport_store import CustomRestStore
from mlflow.tracking._tracking_service.client import TrackingServiceClient
from mlflow.tracking.metric_value_conversion_utils import (
    convert_metric_value_to_float_if_possible,
)
from mlflow.utils import chunk_list
from mlflow.utils.rest_utils import MlflowHostCreds
from mlflow.utils.time import get_current_time_millis
from mlflow.utils.validation import MAX_METRICS_PER_BATCH

from python.fasttrackml.entities.metric import Metric


class FasttrackmlTrackingServiceClient(TrackingServiceClient):

    def __init__(self, tracking_uri):
        super().__init__(tracking_uri)
        self.store_with_context = CustomRestStore(lambda: MlflowHostCreds(self.tracking_uri))

    def log_metric(self, run_id, key, value, timestamp=None, step=None, context=None):
        timestamp = timestamp if timestamp is not None else get_current_time_millis()
        step = step if step is not None else 0
        context = context if context else {}
        metric_value = convert_metric_value_to_float_if_possible(value)
        metric = Metric(key, metric_value, timestamp, step, context)
        
        self.store_with_context.log_metric(run_id, metric)
    
    def log_batch(self, run_id, metrics=()):
        for metrics_batch in chunk_list(metrics, chunk_size=MAX_METRICS_PER_BATCH):
            self.store_with_context.log_batch(run_id=run_id, metrics=metrics_batch)
