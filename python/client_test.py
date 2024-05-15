import os
import subprocess
import time
import uuid
from random import random, uniform

import pytest
from fasttrackml.entities import Metric, Param

from fasttrackml import FasttrackmlClient


@pytest.fixture(scope="session", autouse=True)
def server():
    # Launch the fml server
    process = subprocess.Popen(["fml", "server"])
    yield process
    # Kill the fml server
    process.kill()


@pytest.fixture
def client():
    return FasttrackmlClient("http://localhost:5000")


@pytest.fixture
def run(client, server):
    experiment_id = "0"
    run = client.create_run(experiment_id)
    yield run
    client.set_terminated(run.info.run_id)


def test_log_metric(client, server, run):
    metric_key = str(uuid.uuid4())
    client.log_metric(run.info.run_id, metric_key, random(), context={"context_key": "context_value1"})
    client.log_metric(run.info.run_id, metric_key, random() + 1, context={"context_key": "context_value2"})
    client.log_metric(run.info.run_id, metric_key, random() + 2)

    metric_history = client.get_metric_history(run.info.run_id, metric_key)
    assert metric_history is not None
    assert metric_history[0].key == metric_key


def test_log_batch(client, server, run):
    param_key = str(uuid.uuid4())
    param_value = uniform(0.0, 100.0)
    params = [Param(param_key, param_value)]

    timestamp = int(time.time() * 1000)
    metric_key1 = str(uuid.uuid4())
    metric_key2 = str(uuid.uuid4())
    metrics = [
        Metric(metric_key1, 0.2, timestamp, 1, context={"context_key": "context_value3"}),
        Metric(metric_key2, 0.92, timestamp, 1, context={"context_key": "context_value4"}),
    ]
    client.log_batch(run.info.run_id, metrics=metrics, params=params, synchronous=False)

    metric_keys = [metric_key1, metric_key2]
    metric_histories_df = client.get_metric_histories(run_ids=[run.info.run_id], metric_keys=metric_keys)
    # assert not metric_histories_df.empty
    # assert set(metric_keys).issubset(metric_histories_df['key'].values)
