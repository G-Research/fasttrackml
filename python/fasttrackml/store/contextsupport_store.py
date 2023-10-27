from mlflow import MlflowException
from mlflow.store.tracking.rest_store import RestStore
from mlflow.utils.rest_utils import http_request


class ContextsupportStore(RestStore):

    def __init__(self, host_creds) -> None:
        super().__init__(host_creds)

    def log_metric_with_context(self, run_id, metric):
        context = [{"key": c, "value": metric.context[c]}for c in metric.context]
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
                "context": context,
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

    def log_batch_with_context(self, run_id, metrics):
        metrics_list = []
        for metric in metrics:
            context = [{"key": c, "value": metric.context[c]}for c in metric.context]
            metrics_list.append({
                    "key": metric.key,
                    "value": metric.value,
                    "timestamp": metric.timestamp,
                    "step": metric.step,
                    "context": context,
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

