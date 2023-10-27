from typing import Dict, Optional

from fasttrackml.client import MlflowClientExtend
from fasttrackml.entities.metric_with_context import MetricWithContext
from mlflow.tracking.fluent import _get_or_start_run
from mlflow.utils.time import get_current_time_millis


def log_metric(key: str, value: float, step: Optional[int] = None, context: Optional[dict] = None) -> None:
    run_id = _get_or_start_run().info.run_id
    MlflowClientExtend().log_metric_with_context(run_id, key, value, get_current_time_millis(), step or 0, context or {})

def log_metrics(metrics: Dict[str, float], step: Optional[int] = None, context: Optional[dict] = None) -> None:
    run_id = _get_or_start_run().info.run_id
    timestamp = get_current_time_millis()
    metrics_arr = [MetricWithContext(key, value, timestamp, step or 0, context or {}) for key, value in metrics.items()]
    MlflowClientExtend().log_batch_with_context(run_id=run_id, metrics=metrics_arr)