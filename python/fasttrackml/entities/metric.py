from mlflow.entities import Metric as MlflowMetric


class Metric(MlflowMetric):
    """
    Metric object.
    """

    def __init__(self, key, value, timestamp, step, context=None):
        self._context = context
        MlflowMetric.__init__(self, key, value, timestamp, step)

    @property
    def context(self):
        """Metric context as a Dict."""
        return self._context

    def __hash__(self):
        return hash((self._key, self._value, self._timestamp, self._step, self._context))
