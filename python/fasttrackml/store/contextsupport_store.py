from fasttrackml.protos.metricService_pb2 import LogMetricWrapped, MlflowServiceWrapped
from mlflow.store.tracking.rest_store import RestStore
from mlflow.utils.proto_json_utils import message_to_json
from mlflow.utils.rest_utils import (
    _REST_API_PATH_PREFIX,
    call_endpoint,
    extract_api_info_for_service,
)

_METHOD_TO_INFO = extract_api_info_for_service(MlflowServiceWrapped, _REST_API_PATH_PREFIX)

class ContextsupportStore(RestStore):

    def __init__(self, host_creds) -> None:
        super().__init__(host_creds)

    def log_metric_with_context(self, run_id, metric):
        context_protos = [c.to_proto() for c in metric.context]
        req_body = message_to_json(
            LogMetricWrapped(
                run_uuid=run_id,
                run_id=run_id,
                key=metric.key,
                value=metric.value,
                timestamp=metric.timestamp,
                step=metric.step,
                context=context_protos
            )
        )
        endpoint, method = _METHOD_TO_INFO[LogMetricWrapped]
        response_proto = LogMetricWrapped.Response()
        return call_endpoint(self.get_host_creds(), endpoint, method, req_body, response_proto)

