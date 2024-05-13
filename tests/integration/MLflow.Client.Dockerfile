FROM python:latest

# Install python packages
RUN pip install mlflow psycopg2 boto3
COPY mlflow-setup.py .

