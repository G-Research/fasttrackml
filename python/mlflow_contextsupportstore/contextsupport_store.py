import urllib.parse

from metricWithContext import MetricWithContext
from mlflow.store.tracking.abstract_store import AbstractStore
from mlflow.utils.proto_json_utils import message_to_json
from service_pb2 import LogMetric


class ContextsupportStore(AbstractStore):

    def __init__(self, store_uri = None, artifact_uri = None) -> None:
        path = urllib.parse.urlparse(store_uri).path if store_uri else None
        self.is_plugin = True
        super().__init__(path, artifact_uri)

   
    def log_metric(self, run_id: str, metric: MetricWithContext) -> None:
        req_body = message_to_json(
            LogMetric(
                run_uuid=run_id,
                run_id=run_id,
                key=metric.key,
                value=metric.value,
                timestamp=metric.timestamp,
                step=metric.step,
                context=metric.context,
            )
        )
        self._call_endpoint(LogMetric, req_body)