from parameterized import parameterized
from performance.base import SDKTestBase
from performance.utils import (
    get_baseline,
    write_baseline,
    add_tests_results
)
from performance.sdk.queries import queries
from performance.sdk.utils import (
    collect_runs_data_aim,
    collect_runs_data_mlflow,
    collect_runs_data_fasttrack,
    collect_metrics_data_aim,
    collect_metrics_data_fasttrack,
    collect_metrics_data_mlflow,
)


class TestDataCollectionExecutionTime(SDKTestBase):
    
    @parameterized.expand(queries.items())
    def test_collect_runs_data(self, query_key, query):
        aim_query_execution_time = collect_runs_data_aim(query[0])
        mlflow_query_execution_time = collect_runs_data_mlflow(query[1])
        fasttrack_query_execution_time = collect_runs_data_fasttrack(query[1])
        
        test_name = f'test_collect_runs_data_{query_key}'
        add_tests_results(test_name, aim=aim_query_execution_time, 
                          mlflow=mlflow_query_execution_time, fasttrack=fasttrack_query_execution_time)
        
        baseline = get_baseline(test_name)
        if baseline:
            self.assertInRange(fasttrack_query_execution_time, baseline)
        else:
            write_baseline(test_name, sum([fasttrack_query_execution_time, 
                                           aim_query_execution_time, 
                                           mlflow_query_execution_time])/3)

    @parameterized.expand(queries.items())
    def test_collect_metrics_data(self, query_key, query):
        aim_query_execution_time = collect_metrics_data_aim(query[0])
        mlflow_query_execution_time = collect_metrics_data_mlflow(query[1])
        fasttrack_query_execution_time = collect_metrics_data_fasttrack(query[1])
        test_name = f'test_collect_metrics_data_{query_key}'
        add_tests_results(test_name, aim=aim_query_execution_time, 
                          mlflow=mlflow_query_execution_time, fasttrack=fasttrack_query_execution_time)
        baseline = get_baseline(test_name)
        if baseline:
            self.assertInRange(fasttrack_query_execution_time, baseline)
        else:
            write_baseline(test_name, sum([fasttrack_query_execution_time, 
                                           aim_query_execution_time, 
                                           mlflow_query_execution_time])/3)