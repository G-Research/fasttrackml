import os
import shutil
import tarfile
import time
from pathlib import Path

from aim.sdk.configs import AIM_REPO_NAME
from performance.utils import get_baseline_filename, MLFLOW_CLIENT, FASTTRACK_CLIENT, generate_html_report
import random
from aim import Run

TEST_REPO_PATHS = {
    'real_life_repo': '.aim',
    'generated_repo': '.aim'
}
# AIM_PERFORMANCE_BUCKET_NAME = 'aim-demo-logs'
# AIM_PERFORMANCE_LOG_FILE_NAME = 'performance-logs.tar.gz'


def _init_test_repos():
    
    NUM_RUNS = 10
    POSSIBLE_METRICS = ["accuracy", "precision", "recall", "loss", "F1", "recall"]
    POSSIBLE_BATCH_SIZES = [32, 64, 128, 256, 512]
    POSSIBLE_EXPERIMENT_MODELS = ["VGC", "CNN", "ResNet", "ViT", "U-Net"]
    # initialize aim repo and create runs and add metrics to those runs
    print("SETTING UP AIM")
    for _ in range(NUM_RUNS):
        run = Run()
        hparams_dict = {
            'learning_rate': random.random(),
            'batch_size': random.choice(POSSIBLE_BATCH_SIZES),
            "experiment_model": random.choice(POSSIBLE_EXPERIMENT_MODELS)
        }
        run['hparams'] = hparams_dict
        metrics = random.sample(POSSIBLE_METRICS, 3)
        for _ in range(100):
            for metric in metrics:
                run.track(random.random(), metric)
                
    # initialize mlflow experiment and create runs and add metric to those runs
    print("SETTING UP MLFLOW")
    experiment =  MLFLOW_CLIENT.get_experiment_by_name('test-experiment')
    if experiment:
        EXPERIMENT_ID = experiment.experiment_id
    else:
        EXPERIMENT_ID = MLFLOW_CLIENT.create_experiment('test-experiment')
    
    os.environ['MLFLOW_EXPERIMENT_ID'] = EXPERIMENT_ID
    
    for _ in range(NUM_RUNS):
        run = MLFLOW_CLIENT.create_run(EXPERIMENT_ID)
        RUN_ID = run.info.run_id
        MLFLOW_CLIENT.log_param(RUN_ID, "learning_rate", random.random())
        MLFLOW_CLIENT.log_param(RUN_ID, "batch_size", random.choice(POSSIBLE_BATCH_SIZES))
        MLFLOW_CLIENT.log_param(RUN_ID, "experiment_model", random.choice(POSSIBLE_EXPERIMENT_MODELS))
        
        metrics = random.sample(POSSIBLE_METRICS, 3)
        for _ in range(100):
            for metric in metrics:
                MLFLOW_CLIENT.log_metric(RUN_ID, metric, random.random())
        
    # initialize fasttrack experiment and create runs and add metrics to thos runs
    print("SETTING UP FASTTRACK")
    experiment =  FASTTRACK_CLIENT.get_experiment_by_name('test-experiment')
    if experiment:
        EXPERIMENT_ID = experiment.experiment_id
    else:
        EXPERIMENT_ID = FASTTRACK_CLIENT.create_experiment('test-experiment')
    
    os.environ['FASTTRACK_EXPERIMENT_ID'] = EXPERIMENT_ID
    
    for _ in range(NUM_RUNS):
        run = FASTTRACK_CLIENT.create_run(EXPERIMENT_ID)
        RUN_ID = run.info.run_id
        FASTTRACK_CLIENT.log_param(RUN_ID, "learning_rate", random.random())
        FASTTRACK_CLIENT.log_param(RUN_ID, "batch_size", random.choice(POSSIBLE_BATCH_SIZES))
        FASTTRACK_CLIENT.log_param(RUN_ID, "experiment_model", random.choice(POSSIBLE_EXPERIMENT_MODELS))
        
        metrics = random.sample(POSSIBLE_METRICS, 3)
        for _ in range(100):
            for metric in metrics:
                FASTTRACK_CLIENT.log_metric(RUN_ID, metric, random.random())
        
    os.environ["INITIALIZED"] = 'true'
    print("SETUP COMPLETE")

def _cleanup_test_repo(path):
    shutil.rmtree(path)


def pytest_sessionstart(session):
    if not os.environ.get("INITIALIZED"):
        _init_test_repos()
    time.sleep(10)

def print_current_baseline():
    print('==== CURRENT BASELINE ====')
    with open(get_baseline_filename(), 'r') as f:
        print(f.read())
    print('==========================')

def pytest_unconfigure(config):
    print_current_baseline()
    generate_html_report()

def pytest_sessionfinish(session, exitstatus):
    if os.environ.get('AIM_LOCAL_PERFORMANCE_TEST'):
        for path in TEST_REPO_PATHS.values():
            _cleanup_test_repo(path)
    if os.environ.get(AIM_REPO_NAME):
        del os.environ[AIM_REPO_NAME]