from typing import Dict, Optional

from mlflow.tracking.fluent import _get_or_start_run
from mlflow.utils.time import get_current_time_millis

from .client import FasttrackmlClient
from .entities.metric import Metric


def log_metric(key: str, value: float, step: Optional[int] = None, context: Optional[dict] = None) -> None:
    run_id = _get_or_start_run().info.run_id
    FasttrackmlClient().log_metric(run_id, key, value, get_current_time_millis(), step or 0, context or {})


def log_metrics(metrics: Dict[str, float], step: Optional[int] = None, context: Optional[dict] = None) -> None:
    run_id = _get_or_start_run().info.run_id
    timestamp = get_current_time_millis()
    metrics_arr = [Metric(key, value, timestamp, step or 0, context) for key, value in metrics.items()]
    FasttrackmlClient().log_batch(run_id=run_id, metrics=metrics_arr)
