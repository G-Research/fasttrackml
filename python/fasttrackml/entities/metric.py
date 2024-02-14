from mlflow.entities._mlflow_object import _MLflowObject


class Metric(_MLflowObject):
    """
    Metric object.
    """

    def __init__(self, key, value, timestamp, step, context=None):
        self._key = key
        self._value = value
        self._timestamp = timestamp
        self._step = step
        self._context = context

    @property
    def key(self):
        """String key corresponding to the metric name."""
        return self._key

    @property
    def value(self):
        """Float value of the metric."""
        return self._value

    @property
    def timestamp(self):
        """Metric timestamp as an integer (milliseconds since the Unix epoch)."""
        return self._timestamp

    @property
    def step(self):
        """Integer metric step (x-coordinate)."""
        return self._step

    @property
    def context(self):
        """Metric context as a Dict."""
        return self._context

    def __eq__(self, __o):
        if isinstance(__o, self.__class__):
            return self.__dict__ == __o.__dict__

        return False

    def __hash__(self):
        return hash((self._key, self._value, self._timestamp, self._step, self._context))
