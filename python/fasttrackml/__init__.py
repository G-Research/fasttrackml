import contextlib

import mlflow
from .client import FasttrackmlClient
from mlflow import *

del log_metric, log_metrics
from .fluent import log_metric, log_metrics

__all__ = [name for name in dir() if name in dir(mlflow)]
__all__.append("FasttrackmlClient")

with contextlib.suppress(Exception):
    from mlflow import gateway 
    __all__.append("gateway")
