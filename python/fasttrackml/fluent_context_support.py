from typing import Optional

from fasttrackml.client import MlflowClientExtend
from mlflow.tracking.fluent import _get_or_start_run
from mlflow.utils.time import get_current_time_millis


def log_metric_with_context(key: str, value: float, step: Optional[int] = None, context: Optional[str] = None) -> None:
    run_id = _get_or_start_run().info.run_id
    MlflowClientExtend().log_metric_with_context(run_id, key, value, get_current_time_millis(), step or 0, context or {})