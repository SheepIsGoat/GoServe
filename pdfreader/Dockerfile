# Use an official Python runtime as a parent image
FROM python:3.11

COPY requirements.txt /tmp
RUN pip install --no-cache-dir -r /tmp/requirements.txt

# Set the working directory in the container
WORKDIR /usr/src/app

# Copy the current directory contents into the container at /usr/src/app
COPY main.py /usr/src/app/main.py

# Make port 8000 available to the world outside this container
EXPOSE 8000

# Define environment variable
ENV NAME World

# Run app.py when the container launches
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]