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

# Create the benchmark outputs if they do not exist
mkdir -p benchmark_outputs
# Ensure the created folder is world writable
chmod 777 benchmark_outputs

# Run your specific Docker command
docker-compose up logging_test_mlflow_sqlite
docker-compose up logging_test_mlflow_postgres
docker-compose up logging_test_fasttrack_sqlite
docker-compose up logging_test_fasttrack_postgres
docker-compose up retrieval_test_mlflow_sqlite
docker-compose up retrieval_test_mlflow_postgres
docker-compose up retrieval_test_fasttrack_postgres
docker-compose up retrieval_test_fasttrack_sqlite

docker-compose up performance_benchmark_test

# Shut down all created containers
docker-compose down
