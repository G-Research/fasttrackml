import contextlib

import mlflow
from mlflow import *

del log_metric
from fasttrackml.fluent_context_support import log_metric_with_context as log_metric

__all__ = [name for name in dir() if name in dir(mlflow)]

# `mlflow.gateway` depends on optional dependencies such as pydantic and has version
# restrictions for dependencies. Importing this module fails if they are not installed or
# if invalid versions of these required packages are installed.
with contextlib.suppress(Exception):
    from mlflow import gateway  # noqa: F401

    __all__.append("gateway")
