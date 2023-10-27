from typing import Dict, Optional, Sequence

from fasttrackml._tracking_service.client import TrackingServiceClientExtend
from fasttrackml.entities.metric_with_context import MetricWithContext
from mlflow import MlflowClient
from mlflow.entities import Param, RunTag
from mlflow.tracking._tracking_service import utils


class MlflowClientExtend(MlflowClient):
    def __init__(self, tracking_uri: Optional[str] = None, registry_uri: Optional[str] = None):
        super().__init__(tracking_uri, registry_uri)
        final_tracking_uri = utils._resolve_tracking_uri(tracking_uri)
        self._tracking_client = TrackingServiceClientExtend(final_tracking_uri)

    def log_metric_with_context(
            self,
            run_id: str,
            key: str,
            value: float,
            timestamp: Optional[int] = None,
            step: Optional[int] = None,
            context: Optional[Dict[str, str]] = None,
        ) -> None:
            self._tracking_client.log_metric_with_context(run_id, key, value, timestamp, step, context)
    
    def log_batch_with_context(
            self,
            run_id: str,
            metrics: Sequence[MetricWithContext] = (),
        ) -> None:
            self._tracking_client.log_batch_witch_context(run_id, metrics)