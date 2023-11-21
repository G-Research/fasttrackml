from parameterized import parameterized

from performance.base import SDKTestBase
from performance.utils import (
    get_baseline,
    write_baseline,
    add_tests_results
)
from performance.sdk.queries import queries
from performance.sdk.utils import (
    query_runs_aim,
    query_runs_mlflow, 
    query_runs_fasttrack,
    query_metrics_aim,
    query_metrics_mlflow,
    query_metrics_fasttrack
)


class TestQueryExecutionTime(SDKTestBase):
    @parameterized.expand(queries.items())
    def test_query_runs(self, query_key, query):
        aim_query_execution_time = query_runs_aim(query[0])
        mlflow_query_execution_time = query_runs_mlflow(query[1])
        fasttrack_query_execution_time = query_runs_fasttrack(query[1])
        
        test_name = f'test_query_runs_{query_key}'
        
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
    def test_query_metrics(self, query_key, query):
        aim_query_execution_time = query_metrics_aim(query[0])
        mlflow_query_execution_time = query_metrics_mlflow(query[1])
        fasttrack_query_execution_time = query_metrics_fasttrack(query[1])
        
        test_name = f'test_query_metrics_{query_key}'
        
        add_tests_results(test_name, aim=aim_query_execution_time, 
                          mlflow=mlflow_query_execution_time, fasttrack=fasttrack_query_execution_time)
        baseline = get_baseline(test_name)
        if baseline:
            self.assertInRange(fasttrack_query_execution_time, baseline)
        else:
            write_baseline(test_name, sum([fasttrack_query_execution_time, 
                                           aim_query_execution_time, 
                                           mlflow_query_execution_time])/3)