from aim.sdk import Repo
from aim.sdk.types import QueryReportMode
from aim.web.api.runs.utils import get_run_props
from performance.utils import timing, MLFLOW_CLIENT, FASTTRACK_CLIENT, get_fasttrack_experiment, get_mlflow_experiment
import mlflow

FASTTRACK_EXPERIMENT_ID = get_fasttrack_experiment()
MLFLOW_EXPERIMENT_ID = get_mlflow_experiment()

@timing()
def collect_runs_data_aim(query):
    """
    Test collection of runs data with a particular query on Aim
    Args:
        query (_type_): _description_
    """
    repo = Repo.default_repo()
    runs = repo.query_runs(query, report_mode=QueryReportMode.DISABLED)
    runs_dict = {}
    for run_trace_collection in runs.iter_runs():
        run = run_trace_collection.run
        runs_dict[run.hash] = {
            'params': run[...],
            'traces': run.collect_sequence_info(sequence_types='metric'),
            'props': get_run_props(run)
        }


@timing()
def collect_runs_data_mlflow(query):
    runs = MLFLOW_CLIENT.search_runs(experiment_ids=MLFLOW_EXPERIMENT_ID, filter_string=query)
    runs_dict = {}
    for index, run in enumerate(runs):
        runs_dict[index] = run.data.to_dictionary()
        
@timing()
def collect_runs_data_fasttrack(query):
    runs = FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query)
    runs_dict = {}
    for index, run in enumerate(runs):
        runs_dict[index] = run.data.to_dictionary()

    
@timing()
def collect_metrics_data_aim(query):
    repo = Repo.default_repo()
    runs_dict = {}
    runs = repo.query_metrics(query=query, report_mode=QueryReportMode.DISABLED)
    for run_trace_collection in runs.iter_runs():
        run = None
        traces_list = []
        for trace in run_trace_collection.iter():
            if not run:
                run = run_trace_collection.run
            iters, values = trace.values.sparse_numpy()
            traces_list.append({
                'name': trace.name,
                'context': trace.context.to_dict(),
                'values': values,
                'iters': iters,
                'epochs': trace.epochs.values_numpy(),
                'timestamps': trace.timestamps.values_numpy()
            })
        if run:
            runs_dict[run.hash] = {
                'traces': traces_list,
                'params': run[...],
                'props': get_run_props(run),
            }

@timing()
def collect_metrics_data_mlflow(query):
    runs = MLFLOW_CLIENT.search_runs(experiment_ids=MLFLOW_EXPERIMENT_ID, filter_string=query)
    metrics = []
    for run in runs:
        metrics.append(run.data.metrics)
        
@timing()
def collect_metrics_data_fasttrack(query):
    runs = FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query)
    metrics = []
    for run in runs:
        metrics.append(run.data.metrics)
            

@timing()
def query_runs_aim(query):
    repo = Repo.default_repo()
    runs = list(repo.query_runs(query=query, report_mode=QueryReportMode.DISABLED).iter_runs())
    
@timing()
def query_runs_mlflow(query):
    runs = list(MLFLOW_CLIENT.search_runs(experiment_ids=MLFLOW_EXPERIMENT_ID, filter_string=query))
@timing()
def query_runs_fasttrack(query):
    runs = list(FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query))

@timing()
def query_metrics_aim(query):
    repo = Repo.default_repo()
    metrics = list(repo.query_metrics(query=query, report_mode=QueryReportMode.DISABLED).iter())
    
@timing()
def query_metrics_mlflow(query):
    runs = list(FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query))
    
@timing()
def query_metrics_fasttrack(query):
    runs = list(FASTTRACK_CLIENT.search_runs(experiment_ids=FASTTRACK_EXPERIMENT_ID, filter_string=query))