from aim import Repo

from performance.base import StorageTestBase
from performance.utils import get_baseline, write_baseline, add_tests_results
from performance.storage.utils import iterative_access_metric_values_aim, \
                                      iterative_access_metric_values_mlflow, \
                                      iterative_access_metric_values_fasttrack
                                    


class TestIterativeAccessExecutionTime(StorageTestBase):
    def test_iterative_access(self):
        test_name = 'test_iterative_access'
        repo = Repo.default_repo()
        query = 'metric.acc > 0.6'
        aim_execution_time = iterative_access_metric_values_aim(repo, query)
        mlflow_execution_time = iterative_access_metric_values_mlflow(query)
        fasttrack_execution_time = iterative_access_metric_values_fasttrack(query)
        add_tests_results(test_name, aim=aim_execution_time, 
                          mlflow=mlflow_execution_time, fasttrack=fasttrack_execution_time)
        baseline = get_baseline(test_name)
        if baseline:
            self.assertInRange(fasttrack_execution_time, baseline)
        else:
            write_baseline(test_name, sum([fasttrack_execution_time, aim_execution_time, mlflow_execution_time])/3)