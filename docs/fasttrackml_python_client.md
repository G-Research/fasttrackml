
## FastTrackML Python Client Overview

`fasttrackml` is a Python package that incorporates and enhances the capabilities MLFlow's tracking client package. While offerring the same interface as MLFlow, `fasttrackml` adds some additional methods for convenience and better performance with the FastTrackML tracking server. 


**Enhanced Logging Methods:**

FastTrackML introduces enhanced logging methods that go beyond the standard MLFlow functionality. Notably, the `log_metric` and `log_metrics` methods offer an added layer of flexibility and detail by introducing the concept of context.

-   **`log_metric` Method:**
    
    The `log_metric` method is designed to log individual metrics with context. Beyond the essential parameters of metric name and value, FastTrackML allows users to include an optional `context` parameter. This `context` parameter is a dictionary that enables users to attach additional information to the metric being logged.
    
    ```python
    from fasttrackml import log_metric
    
    # Log a single metric with context
    log_metric("accuracy", 0.85, context={'subset': 'training'})
    ```
    
-   **`log_metrics` Method:**
    
    Similarly, the `log_metrics` method empowers users to log multiple metrics simultaneously while incorporating context. The `context` parameter, when used in conjunction with this method, allows users to provide additional information for each metric in the dictionary. 
    
     ```python
    from fasttrackml import log_metrics 
    
    # Log multiple metrics with context 
    metrics_dict = {'precision': 0.92, 'recall': 0.88} 			
    log_metrics(metrics_dict, context={'subset': 'validation'})
    ```

    
**Extended Metric Retrieval:**

FastTrackML extends the functionality for retrieving metric information by introducing the following methods:

-   **`get_metric_history` Method:**
    
    The `get_metric_history` method retrieves a list of metric objects corresponding to all values logged for a given metric within a specific run. This allows users to explore the detailed history of a metric, including its values, steps, timestamps, and associated context.
    
	 ```python
	from fasttrackml import FasttrackmlClient

	# Create a FasttrackmlClient instance
	client = FasttrackmlClient()

	# Fetch metric history for a specific run and metric
	run_id = "your_run_id"  # Replace with a valid run ID
	metric_key = "accuracy"  # Replace with the desired metric key
	metric_history = client.get_metric_history(run_id, metric_key)

	```
    
-   **`get_metric_histories` Method:**
    
    FastTrackML introduces the `get_metric_histories` method, which is not available in standard MLFlow. This method allows users to retrieve metric histories for multiple runs, metrics, or experiments, providing a convenient way to analyze and compare metric trends across various contexts.
    ```python
	from fasttrackml import FasttrackmlClient

	# Create a FasttrackmlClient instance
	client = FasttrackmlClient()

	# Fetch metric histories for multiple runs and metrics
	run_ids = ["run_id1", "run_id2"]  # Replace with valid run IDs
	metric_keys = ["metric1", "metric2"]  # Replace with desired metric keys
	metric_histories_df = client.get_metric_histories(run_ids=run_ids, metric_keys=metric_keys)
    # Fetch metric histories for multiple runs and metrics with a specific context
	filtered_metric_histories = client.get_metric_histories(run_ids=run_ids, metric_keys=metric_keys, context={"context_key": "context_value1"})
	```
### Example:


   ```python
   import fasttrackml
from fasttrackml import log_metric, log_metrics

# Set MLFlow tracking URI and experiment
fasttrackml.set_tracking_uri("http://localhost:5000")
fasttrackml.set_experiment("my-experiment")

# Log a single metric with context 
log_metric("accuracy", 0.85, step=100, context={'subset': 'training'})

# Log another single metric without context
log_metric("loss", 0.05, step=100)

# Log multiple metrics with context 
metrics_dict_with_context = {'precision': 0.92, 'recall': 0.88}
log_metrics(metrics_dict_with_context, step=200, context={'subset': 'validation'})

# Log another set of metrics without context
metrics_dict_without_context = {'f1_score': 0.89, 'time_elapsed': 120}
log_metrics(metrics_dict_without_context, step=200)
```
