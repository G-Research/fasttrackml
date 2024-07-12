import os
import posixpath
import socket
import subprocess
import time
import uuid
from random import random, uniform

import pytest

from fasttrackml import FasttrackmlClient
from fasttrackml.entities import Metric, Param

LOCALHOST = "127.0.0.1"


@pytest.fixture(scope="session")
def fml_address():
    # Launch the fml server
    port = get_safe_port()
    return f"{LOCALHOST}:{port}"


def get_safe_port():
    """Returns an ephemeral port that is very likely to be free to bind to."""
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.bind((LOCALHOST, 0))
    port = sock.getsockname()[1]
    sock.close()
    return port


@pytest.fixture(scope="session", autouse=True)
def server(fml_address):
    process = subprocess.Popen(["fml", "server"], env={**os.environ, "FML_LISTEN_ADDRESS": f"{fml_address}"})
    yield process
    # Kill the fml server
    time.sleep(3)
    process.kill()


@pytest.fixture
def client(fml_address):
    return FasttrackmlClient(f"http://{fml_address}")


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


def test_log_param(client, server, run):
    # log param string and verify
    param_key = str(uuid.uuid4())
    param_value = str(uuid.uuid4())
    client.log_param(run.info.run_id, param_key, param_value)

    run_response = client.get_run(run.info.run_uuid)
    assert run_response is not None
    assert run_response.data.params[param_key] == param_value

    # log param float
    # TODO verify when get_run works with float value
    param_key = str(uuid.uuid4())
    param_value = uniform(0.0, 100.0)
    client.log_param(run.info.run_id, param_key, param_value)


def test_log_batch(client, server, run):
    param_key = str(uuid.uuid4())
    param_value = uniform(0.0, 100.0)
    params = [Param(param_key, param_value)]

    timestamp = int(time.time() * 1000)
    metric_key1 = str(uuid.uuid4())
    metric_key2 = str(uuid.uuid4())
    metric_key1_value = 0.2
    metric_key2_value = 0.92
    metrics = [
        Metric(metric_key1, metric_key1_value, timestamp, 1, context={"context_key": "context_value3"}),
        Metric(metric_key2, metric_key2_value, timestamp, 1, context={"context_key": "context_value4"}),
    ]
    client.log_batch(run.info.run_id, metrics=metrics, params=params, synchronous=False)

    time.sleep(1)

    metric_keys = [metric_key2]
    metric_histories_df = client.get_metric_histories(run_ids=[run.info.run_id], metric_keys=metric_keys)
    assert metric_histories_df.value[0] == metric_key2_value


def test_log_output(client, server, run):
    # test logging some output directly
    for i in range(100):
        log_data = str(uuid.uuid4()) + "\n" + str(uuid.uuid4())
        assert client.log_output(run.info.run_id, log_data) == None


def test_init_output_logging(client, server, run):
    # test logging some output implicitly
    client.init_output_logging(run.info.run_id)
    for i in range(100):
        log_data = str(uuid.uuid4()) + "\n" + str(uuid.uuid4())
        print(log_data)


def test_log_image(client, server, run):
    # test logging some images
    for i in range(100):
        img_local = posixpath.join(os.path.dirname(__file__), "dice.png")
        assert (
            client.log_image(run.info.run_id, img_local, "images", "These are dice", 0, 640, 480, "png", i, 0) == None
        )
