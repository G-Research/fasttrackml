import json

from mlflow import MlflowException
from mlflow.store.tracking.rest_store import RestStore
from mlflow.utils.rest_utils import http_request


class CustomRestStore(RestStore):

    def __init__(self, host_creds) -> None:
        super().__init__(host_creds)

    def log_metric(self, run_id, metric):
        try:
            json.dumps(metric.context)
        except Exception as e:
            raise MlflowException(f"Failed to serialize object in context: {metric.context}: {str(e)}")
        result = http_request(**{
            "host_creds": self.get_host_creds(),
            "endpoint": "/api/2.0/mlflow/runs/log-metric",
            "method": "POST",
            "json": {
                "run_uuid": run_id,
                "key": metric.key,
                "value": metric.value,
                "timestamp": metric.timestamp,
                "step": metric.step,
                "context": metric.context,
            }
        })
        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

    def log_batch(self, run_id, metrics):
        metrics_list = []
        for metric in metrics:
            metrics_list.append({
                    "key": metric.key,
                    "value": metric.value,
                    "timestamp": metric.timestamp,
                    "step": metric.step,
                    "context": metric.context,
                })
        
        result = http_request(**{
            "host_creds": self.get_host_creds(),
            "endpoint": "/api/2.0/mlflow/runs/log-batch",
            "method": "POST",
            "json": {
                "run_id": run_id,
                "metrics": metrics_list
            }
        })

        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

