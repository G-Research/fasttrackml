import json
import os
import posixpath
import threading
from typing import Dict, Optional, Sequence

import pyarrow as pa
from mlflow import MlflowException
from mlflow.entities import ViewType
from mlflow.exceptions import MlflowException
from mlflow.store.entities.paged_list import PagedList
from mlflow.store.tracking.rest_store import RestStore
from mlflow.tracking import MlflowClient
from mlflow.tracking.fluent import (
    _get_experiment_id,
    get_experiment_by_name,
    search_experiments,
)
from mlflow.utils.async_logging.async_logging_queue import AsyncLoggingQueue
from mlflow.utils.async_logging.run_batch import RunBatch
from mlflow.utils.async_logging.run_operations import RunOperations
from mlflow.utils.rest_utils import http_request
from mlflow.utils.string_utils import is_string_type

from ..entities.metric import Metric


class CustomRestStore(RestStore):

    def __init__(self, host_creds) -> None:
        super().__init__(host_creds)
        self._async_logging_queue = AsyncLoggingQueue(logging_func=self.log_batch)

    def log_metric(self, run_id, metric):
        try:
            json.dumps(metric.context)
        except Exception as e:
            raise MlflowException(f"Failed to serialize object in context: {metric.context}: {str(e)}")
        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/runs/log-metric",
                "method": "POST",
                "json": {
                    "run_id": run_id,
                    "run_uuid": run_id,
                    "key": metric.key,
                    "value": metric.value,
                    "timestamp": metric.timestamp,
                    "step": metric.step,
                    "context": metric.context,
                },
            }
        )
        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

    def log_param(self, run_id, param):
        request_body = {
            "run_id": run_id,
            "run_uuid": run_id,
            "key": param.key,
        }

        if isinstance(param.value, int):
            request_body["value_int"] = param.value
        elif isinstance(param.value, float):
            request_body["value_float"] = param.value
        else:
            request_body["value"] = param.value

        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/runs/log-parameter",
                "method": "POST",
                "json": request_body,
            }
        )
        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

    def log_batch(self, run_id, metrics=[], params=[], tags=[]):
        metrics_list, params_list, tags_list = [], [], []
        for metric in metrics:
            metrics_list.append(
                {
                    "key": metric.key,
                    "value": metric.value,
                    "timestamp": metric.timestamp,
                    "step": metric.step,
                    "context": metric.context,
                }
            )

        for param in params:
            param_partial = {"key": param.key}
            if isinstance(param.value, int):
                param_partial["value_int"] = param.value
            elif isinstance(param.value, float):
                param_partial["value_float"] = param.value
            else:
                param_partial["value"] = param.value
            params_list.append(param_partial)

        for tag in tags:
            tag_list.append(
                {
                    "key": tag.key,
                    "value": tag.value,
                }
            )

        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/runs/log-batch",
                "method": "POST",
                "json": {"run_id": run_id, "metrics": metrics_list, "params": params_list, "tags": tags_list},
            }
        )

        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

    def log_batch_async(self, run_id, metrics=[], params=[], tags=[]) -> RunOperations:
        if not self._async_logging_queue.is_active():
            self._async_logging_queue.activate()

        batch = RunBatch(
            run_id=run_id,
            metrics=metrics,
            params=params,
            tags=tags,
            completion_event=threading.Event(),
        )

        self._async_logging_queue._queue.put(batch)
        operation_future = self._async_logging_queue._batch_status_check_threadpool.submit(
            self._async_logging_queue._wait_for_batch, batch
        )
        return RunOperations(operation_futures=[operation_future])

    def get_metric_history(self, run_id, metric_key, max_results=None, page_token=None):
        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/metrics/get-history",
                "method": "GET",
                "params": {
                    "run_uuid": run_id,
                    "metric_key": metric_key,
                    "max_results": max_results,
                    "page_token": page_token,
                },
            }
        )

        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        js_dict = json.loads(result.text)
        metric_history = [
            Metric(metric["key"], metric["value"], metric["timestamp"], metric["step"], metric["context"])
            for metric in js_dict.get("metrics")
        ]
        next_page_token = js_dict.get("next_page_token")
        return PagedList(metric_history, next_page_token or None)

    def get_metric_histories(
        self,
        experiment_ids: Optional[Sequence[str]] = None,
        run_ids: Optional[Sequence[str]] = None,
        metric_keys: Optional[Sequence[str]] = None,
        index: str = "step",
        run_view_type: int = None,
        max_results: int = 10000000,
        search_all_experiments: bool = False,
        experiment_names: Optional[Sequence[str]] = None,
        context: Optional[Dict[str, object]] = None,
    ):
        if index not in ("step", "timestamp"):
            raise ValueError(f"Unsupported index: {index}. Supported string values are 'step' or 'timestamp'")

        no_exp_ids = experiment_ids is None or len(experiment_ids) == 0
        no_exp_names = experiment_names is None or len(experiment_names) == 0
        no_run_ids = run_ids is None or len(run_ids) == 0
        no_ids_or_names = no_exp_ids and no_exp_names and no_run_ids
        if not no_exp_ids and not no_exp_names:
            raise ValueError(message="Only experiment_ids or experiment_names can be used, but not both")

        if search_all_experiments and no_ids_or_names:
            experiment_ids = [exp.experiment_id for exp in search_experiments(view_type=ViewType.ACTIVE_ONLY)]
        elif no_ids_or_names:
            experiment_ids = _get_experiment_id()
        elif not no_exp_names:
            experiments = [get_experiment_by_name(n) for n in experiment_names if n is not None]
            experiment_ids = [e.experiment_id for e in experiments if e is not None]

        if isinstance(experiment_ids, int) or is_string_type(experiment_ids):
            experiment_ids = [experiment_ids]
        if is_string_type(run_ids):
            run_ids = [run_ids]
        if is_string_type(metric_keys):
            metric_keys = [metric_keys]

        # Using an internal function as the linter doesn't like assigning a lambda, and inlining the
        # full thing is a mess
        result = http_request(
            host_creds=self.get_host_creds(),
            endpoint="/api/2.0/mlflow/metrics/get-histories",
            method="POST",
            json={
                "experiment_ids": experiment_ids,
                "run_ids": run_ids,
                "metric_keys": metric_keys,
                "run_view_type": ViewType.to_string(run_view_type).upper(),
                "max_results": max_results,
                "context": context,
            },
            stream=True,
        )

        if result.status_code != 200:
            result = result.json()
            if "error_code" in result:
                raise MlflowException(
                    message=result["message"],
                    error_code=result["error_code"],
                )

        with pa.ipc.open_stream(result.raw) as reader:
            return reader.read_pandas().set_index(["run_id", "key", index])

    def log_output(self, run_id, data):
        request_body = {
            "run_id": run_id,
            "data": data,
        }

        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/runs/log-output",
                "method": "POST",
                "json": request_body,
            }
        )
        if result.status_code != 200:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result

    def log_image(
        self,
        run_id: str,
        name: str,
        filename: str,
        artifact_path: str,
        caption: str,
        index: int,
        width: int,
        height: int,
        format: str,
        step: int,
        iter: int,
    ):
        storage_path = posixpath.join(artifact_path, os.path.basename(filename))
        request_body = {
            "run_id": run_id,
            "name": name,
            "blob_uri": storage_path,
            "caption": caption,
            "index": index,
            "width": width,
            "height": height,
            "format": format,
            "step": step,
            "iter": iter,
        }
        result = http_request(
            **{
                "host_creds": self.get_host_creds(),
                "endpoint": "/api/2.0/mlflow/runs/log-artifact",
                "method": "POST",
                "json": request_body,
            }
        )
        if result.status_code != 201:
            result = result.json()
        if "error_code" in result:
            raise MlflowException(
                message=result["message"],
                error_code=result["error_code"],
            )
        return result
