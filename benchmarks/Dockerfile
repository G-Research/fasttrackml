# Use an official Python runtime as a parent image
FROM python:3.8

# Expose default mlflow port
EXPOSE 5000

# Install any needed packages specified in requirements.txt
RUN pip install psycopg2-binary fasttrackml mlflow

CMD [ "mlflow" , "server" ]
