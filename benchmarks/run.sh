#!/bin/bash

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Run your specific Docker command
docker-compose up throughput_test_mlflow_sqlite
docker-compose up throughput_test_mlflow_postgres
docker-compose up throughput_test_fasttrack_sqlite
docker-compose up throughput_test_fasttrack_postgres
docker-compose up retreival_test_mlflow_sqlite
docker-compose up retreival_test_mlflow_postgres
docker-compose up retreival_test_fasttrack_postgres
docker-compose up retreival_test_fasttrack_sqlite

docker-compose up performance_benchmark_test

# Shut down all created containers
docker-compose down
