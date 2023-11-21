from aim import Repo
from aim.sdk.configs import get_aim_repo_name
from aim.sdk.types import QueryReportMode

from performance.utils import timing, MLFLOW_CLIENT, FASTTRACK_CLIENT, get_fasttrack_experiment, get_mlflow_experiment

FASTTRACK_EXPERIMENT_ID = get_fasttrack_experiment()
MLFLOW_EXPERIMENT_ID = get_mlflow_experiment()


@timing()
def random_access_metric_values_aim(repo: Repo, query, density):
    traces = repo.query_metrics(query=query, report_mode=QueryReportMode.DISABLED)
    for trace in traces.iter():
        values = trace.values
        values_length = len(values)
        step = len(values)//density

        accessed_values = []
        for i in range(0, values_length, step):
            accessed_values.append(trace.values[i])

@timing()  
def random_access_metric_values_mlflow(query):
    runs = MLFLOW_CLIENT.search_runs(experiment_ids=MLFLOW_EXPERIMENT_ID, filter_string=query)
    metrics = [ run.data.metrics for run in runs]
    metric_data = {}
    for metric in metrics:
        for k, v in metric.items():
            metric_data[k] = v if not metric_data[k] else (metric_data[k] + v)/2
            
@timing() 
def random_access_metric_value_fasttrack(query):
    runs = FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query)
    metrics = [ run.data.metrics for run in runs]
    metric_data = {}
    for metric in metrics:
        for k, v in metric.items():
            metric_data[k] = v if not metric_data[k] else (metric_data[k] + v)/2

@timing()
def iterative_access_metric_values_aim(repo:Repo, query):
    traces = repo.query_metrics(query=query, report_mode=QueryReportMode.DISABLED)
    for trace in traces.iter():
        _ = trace.values.values_numpy()

@timing()
def iterative_access_metric_values_mlflow(query):
    runs = MLFLOW_CLIENT.search_runs(experiment_ids=MLFLOW_EXPERIMENT_ID, filter_string=query)
    metrics = [ run.data.metrics for run in runs]
    for metric in metrics:
        _ = metric.values()
             
@timing()
def iterative_access_metric_values_fasttrack(query):
    runs = FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query)
    metrics = [ run.data.metrics for run in runs]
    for metric in metrics:
        _ = metric.values()