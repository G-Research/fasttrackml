# Search Syntax Documentation
## Introduction

This document provides an overview of the search syntax that users can utilize to filter metrics and runs.


- [Search Runs](#search-runs)
- [Search Metrics](#search-metrics)
- [Operations](#operations)
  - [String operations](#string-operations)
  - [Numeric operations](#numeric-operations)
  - [Boolean operations](#boolean-operations)
  - [Logical operations](#logical-operations)
- [Search run examples](#search-run-examples)
  - [Example with run.name (string)](#example-with-runname-string)
  - [Example with run.duration (numeric)](#example-with-runduration-numeric)
  - [Example with run.archived (boolean)](#example-with-runarchived-boolean)
  - [Run parameters](#run-parameters)
  - [Filtering Runs with Unset Attributes](#filtering-runs-with-unset-attributes)
  - [Filter Runs using Regular Expressions](#filter-runs-using-regular-expressions)
  - [Complex query for run search](#complex-query-for-run-search)
- [Search metrics examples](#search-metrics-examples)
  - [Example with metric.name (string)](#example-with-metricname-string)
  - [Example with metric.last (numeric)](#example-with-metriclast-numeric)
  - [Filter Metrics by run](#filter-metrics-by-run)
  - [Complex query for metric search](#complex-query-for-metric-search)


## Search Runs
You can filter the runs using the following run attributes:
| Property           | Description                                         | Type             |
| ------------------ | --------------------------------------------------- | ---------------- |
| ```name```         | Run name                                            | ```string```     |
| ```hash```         | Run hash                                            | ```string```     |
| ```experiment```   | Experiment name                                     | ```string```     |
| ```tags```         | List of run tags                                    | ```dictionary``` |
| ```archived```     | True if run is archived, otherwise False            | ```boolean```    |
| ```active```       | True if run is active(in progress), otherwise False | ```boolean```    |
| ```duration```     | Run duration in seconds                             | ```numeric```    |
| ```created_at```   | Run creation datetime                               | ```numeric```    |
| ```finalized_at``` | Run end datetime                                    | ```numeric```    |
| ```metrics```      | Set of run metrics                                  | ```dictionary``` |

## Search Metrics
You can filter the metrics using the following metric attributes:
| Property         | Type          |
| ---------------- | ------------- |
| ```name```       | ```string```  |
| ```last```       | ```numeric``` |
| ```last_step```  | ```numeric``` |
| ```first_step``` | ```numeric``` |

## Operations

### String operations
For the ```string``` attributes you can use the following comparing operator:
- ``` == ```
- ``` != ```
- ``` in ```
- ``` .startswith() ```
- ``` .endswith() ```

### Numeric operations
For the ```numeric``` attributes you can use the following comparing operator:
- ``` == ```
- ``` != ```
- ``` > ```
- ``` >= ```
- ``` < ```
- ``` <= ```

### Boolean operations
For the ```boolean``` attributes you can use the following comparing operator:
- ``` == ```
- ``` != ```

### Logical operations
You can create complex search queries combining multiple conditions with logical operators.
- ``` and ```
- ``` or ```
- ``` not ```



## Search run examples

### Example with ```run.name``` (string)
Select only the runs where the name exactly matches "TestRun1"

```python
run.name == "TestRun1"
```

Select only the runs where the name is different from "TestRun1"
```python
run.name != "TestRun1"
```

Select only the runs where "Run1" is contained in the run name
```python
"Run1" in run.name
```

Select only the runs where the name starts with "Test"
```python
run.name.startswith('Test')
```

Select only the runs where the name ends with "Run1"
```python
run.name.endswith('Run1')
```

### Example with ```run.duration``` (numeric)

Select only the runs where the duration is exactly 111111111

```python
run.duration == 111111111
```


Select only the runs where the duration is not 111111111
```python
run.duration != 111111111
```

Select only the runs where the duration is greater than 111111111
```python
run.duration > 111111111
```

Select only the runs where the duration is greater or equal to 111111111
```python
run.duration >= 111111111
```

Select only the runs where the duration is less than 111111111
```python
run.duration < 111111111
```

Select only the runs where the duration is less or equal to 111111111
```python
run.duration <= 111111111
```

### Example with ```run.archived``` (boolean)
Select only the runs where the archived attribute is true
```python
run.archived == True
```

Select only the runs where the archived attribute is not true
```python
run.archived != True
```

### Run parameters
Run parameters could be accessed via attributes.
![FastTrackML Run List, param filter](images/runs_param_filter.png)

### Filtering Runs with Unset Attributes

To filter runs based on whether an attribute is not set, you can use the following syntax:

```python
run.attribute is None
```
This expression will return runs for which the specified attribute is not defined.

Showing all the runs
![FastTrackML Run List](images/runs.png)

Showing only the runs where param1 is not set
![FastTrackML Run List of not set param](images/runs_none_param.png)

### Filter Runs using Regular Expressions
- ``` .match() ```
- ``` .search() ```

Match finds an exact match at the beginning of a string.
![FastTrackML Run filter using regular expression match](images/runs_regular_expression_match.png)

Search looks for a pattern anywhere in the string.
![FastTrackML Run filter using regular expression match](images/runs_regular_expression_search.png)

### Complex query for run search
The query selects the runs that meet the following conditions:


- run.archived can be either True or False.
- The duration of run must be greater than 0.
- The run has to contain a metric named 'TestMetric' and is value of last  must be greater than 2.5.
- The name of run should not end with '4'.
  
```python
(run.archived == True or run.archived == False) and run.duration > 0 and run.metrics['TestMetric'].last > 2.5 and not run.name.endswith('4')
```

## Search metrics examples

### Example with ```metric.name``` (string)
Select only the metrics where the name exactly matches "TestMetric1"
```python
metric.name == "TestRun1"
```

Select only the metrics where the name is different from "TestMetric1"
```python
metric.name != "TestRun1"
```

Select only the metrics where "Run1" is contained in the run name
```python
"Metric1" in metric.name
```

Select only the metrics where the name starts with "Test"
```python
metric.name.startswith('Test')
```

Select only the metrics where the name ends with "Metric1"
```python
metric.name.endswith('Metric1')
```


### Example with ```metric.last``` (numeric)

Select only the metrics where the last value is exactly 1.1

```python
metric.last == 1.1
```

Select only the metrics where the last value is not 1.1
```python
metric.last != 1.1
```

Select only the metrics where the last value is greater than 1.1
```python
metric.last > 1.1
```

Select only the metrics where the duration is greater or equal to 1.1
```python
metric.last >= 1.1
```

Select only the metrics where the last value is less than 1.1
```python
metric.last < 1.1
```

Select only the metrics where the last value is less or equal to 1.1
```python
metric.last <= 1.1
```

### Filter Metrics by run
You can also filter the metrics by combining  metric attributes with run attributes.

Showing the metrics with the last value greater than 6 belonging to a run with the name that starts with marvelous.

![FastTrackML Metric List filter by metric and run attributes](images/metrics_filter_by_run.png)

### Complex query for metric search
The query selects the metrics that meet the following conditions:

- The metric.name field must be "TestMetric1" or "TestMetric2."
- The metric.last_step field must be greater than or equal to 1.
- The run.name field must either end with "2" or start with "TestRun1."
- The metric.last field must be less than 1.6.
- The run.duration must be greater than 0.
  
```python
((metric.name == "TestMetric1") or (metric.name == "TestMetric2")) and metric.last_step >= 1 and (run.name.endswith("2") or re.match("TestRun1", run.name)) and (metric.last < 1.6) and run.duration > 0
```