import json

from fasttrackml.entities.metric import Metric
from mlflow import MlflowException
from mlflow.store.entities.paged_list import PagedList
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

    def get_metric_history(self, run_id, metric_key, max_results=None, page_token=None):
        result = http_request(**{
            "host_creds": self.get_host_creds(),
            "endpoint": "/api/2.0/mlflow/metrics/get-history",
            "method": "GET",
            "params": {
                "run_uuid": run_id,
                "metric_key": metric_key,
                "max_results": max_results,
                "page_token": page_token,
            }
        })
        
        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        js_dict = json.loads(result.text)
        metric_history = [Metric(metric["key"], metric["value"], metric["timestamp"], metric["step"], metric["context"]) for metric in js_dict.get("metrics")]
        next_page_token = js_dict.get("next_page_token")
        return PagedList(metric_history, next_page_token or None)