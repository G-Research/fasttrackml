import unittest
import os



class TestBase(unittest.TestCase):
    @staticmethod
    def assertInRange(new, base, deviation=0.5):
        failure_message = f'execution time {new}  is greater than of allowed range {base}'
        assert new <= base, failure_message


class SDKTestBase(TestBase):
    @classmethod
    def setUpClass(cls) -> None:
        super().setUpClass()
        # os.environ[AIM_REPO_NAME] = TEST_REPO_PATHS['real_life_repo']

    @classmethod
    def tearDownClass(cls) -> None:
        # del os.environ[AIM_REPO_NAME]
        super().tearDownClass()


class StorageTestBase(TestBase):
    @classmethod
    def setUpClass(cls) -> None:
        super().setUpClass()
        # os.environ[AIM_REPO_NAME] = TEST_REPO_PATHS['generated_repo']

    @classmethod
    def tearDownClass(cls) -> None:
        # del os.environ[AIM_REPO_NAME]
        super().tearDownClass()