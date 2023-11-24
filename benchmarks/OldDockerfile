# Use an official Python runtime as a parent image
FROM python:3.8

# Set the working directory to /app
WORKDIR /src

# # Copy the current directory contents into the container at /app
# COPY . /src

# Install any needed packages specified in requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

# Run app.py when the container launches
CMD ["pytest", "performance"]