# Use Python 3.11 slim image
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Copy requirements first for better caching
COPY requirements.txt .

# Install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .
COPY parser.py .

# Create non-root user for security
RUN useradd -m -u 1000 appuser && chown -R appuser:appuser /app
USER appuser

# Expose port
EXPOSE 8081

# Set environment variables
ENV PORT=8081
ENV SERVICE_NAME=parser-service
ENV PYTHONUNBUFFERED=1

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8081/health')"

# Run application with gunicorn
CMD ["gunicorn", "--bind", "0.0.0.0:8081", "--workers", "2", "--timeout", "30", "--access-logfile", "-", "--error-logfile", "-", "app:app"]
