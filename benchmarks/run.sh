#!/bin/bash

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker first."
    exit 1
fi



# Recreate the benchmark outputs
rm -rf benchmark_outputs
mkdir -p benchmark_outputs
# Ensure the created folder is world writable
chmod 777 benchmark_outputs

# Run the tests containers
for app in "mlflow" "fasttrack"; do
    for db in "sqlite" "postgres"; do
	for test in "logging" "retrieval"; do
	    docker compose down
        docker volume prune -af
	    command="docker compose up ${test}_test_${app}_${db}"
	    echo "Executing command: $command"
	    eval $command
	done
    done
done

# generate the report
docker compose up generate_report
docker compose down
docker volume  prune -af
