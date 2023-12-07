"""
Code edited from https://stable-baselines3.readthedocs.io/en/master/modules/ppo.html
"""


import sys
from typing import Any, Dict, Tuple, Union

import mlflow
import numpy as np
import torch
from stable_baselines3 import PPO
from stable_baselines3.common.callbacks import BaseCallback
from stable_baselines3.common.env_util import make_vec_env
from stable_baselines3.common.logger import HumanOutputFormat, KVWriter, Logger
from stable_baselines3.common.vec_env import VecVideoRecorder
from tqdm import tqdm


class TqdmCallback(BaseCallback):
    def __init__(self):
        super().__init__()
        self.progress_bar = None

    def _on_training_start(self):
        self.progress_bar = tqdm(total=self.locals["total_timesteps"])

    def _on_step(self):
        self.progress_bar.update(3)
        return True

    def _on_training_end(self):
        self.progress_bar.close()
        self.progress_bar = None


class MLflowOutputFormat(KVWriter):
    """
    Dumps key/value pairs into MLflow's numeric format.
    """

    def write(
        self,
        key_values: Dict[str, Any],
        key_excluded: Dict[str, Union[str, Tuple[str, ...]]],
        step: int = 0,
    ) -> None:
        for (key, value), (_, excluded) in zip(sorted(key_values.items()), sorted(key_excluded.items())):
            if excluded is not None and "mlflow" in excluded:
                continue

            if isinstance(value, np.ScalarType):
                if not isinstance(value, str):
                    mlflow.log_metric(key, value, step)

        device = torch.device("cuda" if torch.cuda.is_available() else "cpu")

        allocated = round(torch.cuda.memory_allocated(device=device) / 1024**3, 1)
        cached = round(torch.cuda.memory_reserved(device=device) / 1024**3, 1)
        utilization = torch.cuda.utilization(device=device)

        mlflow.log_metric("GPU Allocated (GB)", allocated)
        mlflow.log_metric("GPU Cached (GB)", cached)
        mlflow.log_metric("GPU Utilization (%)", utilization)


loggers = Logger(
    folder=None,
    output_formats=[HumanOutputFormat(sys.stdout), MLflowOutputFormat()],
)

mlflow.set_tracking_uri("http://localhost:5000")
mlflow.set_experiment("cartpole-ppo-v1")

vec_env = make_vec_env("CartPole-v1", n_envs=4)
# record steps at intervals increaseing with loge
vec_env = VecVideoRecorder(vec_env, "vidcap", record_video_trigger=lambda x: (x & (x - 1)) == 0)

model = PPO("MlpPolicy", vec_env, verbose=1)
model.set_logger(loggers)

with mlflow.start_run():
    model.learn(total_timesteps=25000, callback=TqdmCallback(), log_interval=1)
    model.save("ckpt/ppo_cartpole")

    print("Done Training!!")

vec_env.close()
