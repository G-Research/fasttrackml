FROM python:latest

# Install python packages
RUN pip install "mlflow==${MLFLOW_VERSION:-2.13.0}" psycopg2 boto3
COPY mlflow-setup.py .

