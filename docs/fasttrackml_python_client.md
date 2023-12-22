
## FastTrackML Python Client Overview

`fasttrackml` is a Python package that extends MLFlow's capabilities by providing additional methods.


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