from fasttrackml.protos.metricService_pb2 import MetricContext as ProtoMetricContext
from mlflow.entities._mlflow_object import _MLflowObject


class MetricContext(_MLflowObject):
    """Context object associated with a metric."""

    def __init__(self, key, value):
        self._key = key
        self._value = value

    def __eq__(self, other):
        if type(other) is type(self):
            return self.__dict__ == other.__dict__
        return False

    @property
    def key(self):
        """String key of the context."""
        return self._key

    @property
    def value(self):
        """String value of the context."""
        return self._value

    def to_proto(self):
        param = ProtoMetricContext()
        param.key = self.key
        param.value = self.value
        return param

    @classmethod
    def from_proto(cls, proto):
        return cls(proto.key, proto.value)
