from typing import Dict, Optional, Sequence

from fasttrackml._tracking_service.client import FasttrackmlTrackingServiceClient
from fasttrackml.entities.metric import Metric
from mlflow import MlflowClient
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
            context: Optional[Dict[str, str]] = None,
        ) -> None:
            self._tracking_client.log_metric(run_id, key, value, timestamp, step, context)
    
    def log_batch(
            self,
            run_id: str,
            metrics: Sequence[Metric] = (),
        ) -> None:
            self._tracking_client.log_batch(run_id, metrics)