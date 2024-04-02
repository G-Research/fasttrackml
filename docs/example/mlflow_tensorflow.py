import matplotlib
import matplotlib.pyplot as plt
import mlflow
import pandas as pd
import tensorflow as tf
from sklearn.datasets import make_circles

matplotlib.use("Agg")  # Needed for non-GUI environments

mlflow.set_tracking_uri("http://localhost:5000/")

mlflow.tensorflow.autolog()

# Generate randomized experiments for testing
random_number_samples = tf.random.uniform(shape=(), minval=100, maxval=10000)
random_number_epochs = tf.random.uniform(shape=(), minval=10, maxval=100)

mlflow.set_experiment(f"mlflow_tensorflow_n={random_number_samples}_epochs={random_number_epochs}")

N_SAMPLES = random_number_samples.numpy().astype(int)

# Make some circles
x, y = make_circles(N_SAMPLES, noise=0.03, random_state=42)

# Visualize the data
circles = pd.DataFrame({"x0": x[:, 0], "x1": x[:, 1], "label": y})

# Plot the data
plt.scatter(x[:, 0], x[:, 1])

# Set the random seed
tf.random.set_seed(42)

# 1. Create the model using the Sequential API
model_1 = tf.keras.Sequential([tf.keras.layers.Dense(10, activation="relu"), tf.keras.layers.Dense(1)])

# 2. Compile the model
model_1.compile(
    loss=tf.keras.losses.BinaryCrossentropy(),
    optimizer=tf.keras.optimizers.Adam(learning_rate=0.01),
    metrics=["accuracy"],
)

# 3. Fit the model
model_1.fit(x, y, epochs=random_number_epochs.numpy().astype(int))
