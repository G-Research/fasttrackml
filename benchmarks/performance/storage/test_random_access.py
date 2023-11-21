from parameterized import parameterized

from aim import Repo

from performance.base import StorageTestBase
from performance.utils import get_baseline, write_baseline, add_tests_results
from performance.storage.utils import random_access_metric_values_aim, \
                                      random_access_metric_values_mlflow, \
                                      random_access_metric_value_fasttrack


class TestRandomAccess(StorageTestBase):
    @parameterized.expand({0: 50, 1: 250, 2: 500}.items())
    def test_random_access(self, test_key, density):
        test_name = f'test_random_access_{test_key}'
        repo = Repo.default_repo()
        query = 'metric.acc < 0.6'
        aim_execution_time = random_access_metric_values_aim(repo, query, density)
        mlflow_execution_time = random_access_metric_values_mlflow(query)
        fasttrack_execution_time = random_access_metric_value_fasttrack(query)
        
        add_tests_results(test_name, aim=aim_execution_time, 
                          mlflow=mlflow_execution_time, fasttrack=fasttrack_execution_time)
        
        baseline = get_baseline(test_name)
        if baseline:
            self.assertInRange(fasttrack_execution_time, baseline)
        else:
            write_baseline(test_name, sum([fasttrack_execution_time, aim_execution_time, mlflow_execution_time])/3)