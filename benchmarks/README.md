# FastTrackML Benchmark Suite

## Introduction

FastTrackML Benchmark Suite is a project designed to provide a comprehensive and user-friendly performance benchmarking system for FastTrackML, with a specific focus on comparing its capabilities with other popular machine learning parameter servers, such as MLflow. This documentation aims to guide users on how to use FastTrackML Performance Benchmark effectively.

## Table of Contents

- [FastTrackML Benchmark Suite](#fasttrackml-benchmark-suite)
  - [Introduction](#introduction)
  - [Table of Contents](#table-of-contents)
  - [2. Getting Started ](#2-getting-started-)
  - [3. Usage ](#3-usage-)
    - [Benchmarking Performance ](#benchmarking-performance-)
    - [Results ](#results-)

## 2. Getting Started <a name="getting-started"></a>

To run the performance benchmark ensure you have Docker and Docker Compose installed and run the following command:

```bash
./run.sh
```

## 3. Usage <a name="usage"></a>

### Benchmarking Performance <a name="benchmarking-performance"></a>

FastTrackML benchmark suite allows you to test the performance of the FastTrackML project and compare it to MLFlow through the REST API. We do this by orchestrating 4 containers to to be tested:
- FastTrackML with sqlite
- FastTrackML with postgres
- MLflow with sqlite
- MLflow with postgres

We then perform 2 categories of API benchmark tests on them using the K6 benchmarking tool. The categories of tests are:
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

*Note* These tests in isolation will generate csv report files, but will not generate report images. To generate a report image you will have to run all the tests on all 4 instances then use the `generateReports.py` script to generate the `performanceReport.png` image

*Note* For the performance tests to work you must have a `\benchmark_outputs` folder in this directory. If you are running the pefromance benchmarks without the `run.sh` script you will have to create this folder manually.

### Results <a name="comparing-with-mlflow"></a>

FastTrackML Performance Benchmark is designed to perform benchmark tests on both MLflow and FastTrackML:

![Performance Report](performanceReport.png)
FastTrackML offers the same functionality as MLflow but implements performance optimizations behind the scene to improve overall performance.


Thank you for choosing FastTrackML Performance Tracker! We hope this documentation helps you effectively track and manage your machine learning experiments and compare it with other parameter servers like MLflow.