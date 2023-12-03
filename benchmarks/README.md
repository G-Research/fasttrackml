# FastTrack ML Benchmark Suite

## Introduction

FastTrack ML Benchmark Suite is a project designed to provide a comprehensive and user-friendly performance benchmarking system for FastTrack ML, with a specific focus on comparing its capabilities with other popular machine learning parameter servers, such as MLflow. This documentation aims to guide users on how to use FastTrack ML Performance Benchmark effectively.

## Table of Contents

- [FastTrack ML Benchmark Suite](#fasttrack-ml-benchmark-suite)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [2. Getting Started ](#2-getting-started-)
  - [3. Usage ](#3-usage-)
    - [Benchmarking Performance ](#benchmarking-performance-)
    - [Results ](#results-)

## 2. Getting Started <a name="getting-started"></a>

To run the performance benchmark ensure you have docker and docker compose installed and run the following command:

```bash
./run.sh
```

## 3. Usage <a name="usage"></a>

### Benchmarking Performance <a name="benchmarking-performance"></a>

FastTrack ML benchmark suite allows you to test the performance of the FastTrackML project and compare it to MLFlow through the REST API. We do this by orchestrating 4 containers to to be tested:
- FasttrackML with sqlite
- FasttrackML with postgres
- MLflow with sqlite
- MLflow with postgres

We then perform 2 categories of API benchmark tests on them using the K6 benchmarkign tool. The categories of tests are:
- Logging (throughput)
- Retrieval

You run tests on any of these platforms in isolation for example:

1. To test FastTrackML postgres in isolation:

```bash
docker-compose up logging_test_fasttrack_postgres
```

1. To test performance of MLflow sqlite:

```bash
docker-compose up retreival_test_mlflow_sqlite
```

*Note* These tests in isolation will generate csv report files, but will not generate report images. To generate a report image you will have to run all the tests on all 4 instances then use the `generateReports.py` script to generate the `perfromanceReport.png` image

*Note* For the performance tests to work you must have a `\benchmark_outputs` folder in this directory. If you are running the pefromance benchmarks without the `run.sh` script you will have to create this folder manually.

### Results <a name="comparing-with-mlflow"></a>

FastTrack ML Performance Benchmark is designed to perform benchmark tests on both MLflow and FasttrackML:

![Performance Report](performanceReport.png)
FastTrack ML offers the same functionality as MLflow but implements performance optimizations behind the scene to improve overall performance.


Thank you for choosing FastTrack ML Performance Tracker! We hope this documentation helps you effectively track and manage your machine learning experiments and compare it with other parameter servers like MLflow.