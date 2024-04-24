#!/bin/bash

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker compose &> /dev/null; then
    echo "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create the benchmark outputs if they do not exist
mkdir -p benchmark_outputs
# Ensure the created folder is world writable
chmod 777 benchmark_outputs

# Run the tests containers
for app in "mlflow" "fasttrack"; do
    for db in "sqlite" "postgres"; do
	for test in "logging" "retrieval"; do
	    docker compose down
	    command="docker compose up ${test}_test_${app}_${db}"
	    echo "Executing command: $command"
	    eval $command
	done
    done
done

# generate the report
docker build -t fml-benchmark:generator . && docker run -d -v .:/work -w /work fml-benchmark:generator
