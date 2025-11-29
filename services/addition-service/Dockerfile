FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .

# Expose the service port
EXPOSE 8082

# Run with gunicorn for production
CMD ["gunicorn", "--bind", "0.0.0.0:8082", "--workers", "2", "--threads", "4", "app:app"]
