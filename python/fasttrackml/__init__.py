import contextlib

import mlflow
from mlflow import *

del log_metric, log_metrics
from fasttrackml.fluent import log_metric, log_metrics

__all__ = [name for name in dir() if name in dir(mlflow)]

with contextlib.suppress(Exception):
    from mlflow import gateway 
    __all__.append("gateway")
