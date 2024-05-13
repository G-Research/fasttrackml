FROM python:latest

# Install python packages
RUN pip install mlflow psycopg2 boto3
COPY mlflow-setup.py .
CMD mlflow server --backend-store-uri ${BACKEND_STORE_URI} --host 0.0.0.0 --port 5000

