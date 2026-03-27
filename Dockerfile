# Use an official Python runtime as a parent image
FROM python:3.12-slim

# Set environment variables
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    # Prevents Python from writing pyc files to disc (equivalent to PYTHONDONTWRITEBYTECODE=1)
    # Also prevents buffering so that logs appear in real time

# Set work directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
        gcc \
        curl \
        && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY pyproject.toml ./
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir .[api,db] \
    # Install cloud provider dependencies conditionally (they are optional in pyproject.toml but we want them for prod)
    && pip install --no-cache-dir boto3 azure-mgmt-compute google-api-python-client

# Copy the source code
COPY . .

# Create non-root user for security
RUN useradd --create-home --shell /bin/bash novabackup
RUN chown -R novabackup:novabackup /app
USER novabackup

# Expose the port the app runs on
EXPOSE 8000

# Health check endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8000/docs || exit 1

# Run the application
CMD ["uvicorn", "novabackup.api:get_app", "--host", "0.0.0.0", "--port", "8000"]