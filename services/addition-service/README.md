# Addition Service

The Addition Service is a Python/Flask microservice that performs addition operations as part of the Calculator-as-a-Service (CaaS) architecture.

## Overview

- **Language:** Python 3.12 with Flask
- **Port:** 8082
- **Operator:** `+` (addition)

## API Endpoints

### POST /execute

Execute an addition operation.

**Request:**
```json
{
  "operation": {
    "type": "operation",
    "operator": "+",
    "left": 5,
    "right": 3
  }
}
```

**Response:**
```json
{
  "result": 8,
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "addition-service-7d9f8b6c5-xk2m9",
      "service": "addition-service",
      "operation": "+",
      "operands": {
        "left": 5,
        "right": 3
      },
      "result": 8,
      "delegations": {
        "left": null,
        "right": null
      },
      "duration": 2
    }
  ]
}
```

### GET /health

Liveness probe endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "addition-service"
}
```

### GET /ready

Readiness probe endpoint.

**Response:**
```json
{
  "status": "ready",
  "service": "addition-service"
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Name of this service | `addition-service` |
| `PORT` | Port to listen on | `8082` |
| `SUBTRACTION_SERVICE_URL` | URL of subtraction service | `http://subtraction-service:8083` |
| `MULTIPLICATION_SERVICE_URL` | URL of multiplication service | `http://multiplication-service:8084` |
| `DIVISION_SERVICE_URL` | URL of division service | `http://division-service:8086` |

## Running Locally

### Development Mode

```bash
# Install dependencies
pip install -r requirements.txt

# Run the service
python app.py
```

### Production Mode

```bash
# Install dependencies
pip install -r requirements.txt

# Run with gunicorn
gunicorn --bind 0.0.0.0:8082 --workers 2 --threads 4 app:app
```

### Docker

```bash
# Build the image
docker build -t calculator/addition-service:latest .

# Run the container
docker run -p 8082:8082 calculator/addition-service:latest
```

## Kubernetes Deployment

```bash
# Apply the deployment and service
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## Delegation

The Addition Service can handle nested operations by delegating to other operation services. For example, when processing `5 + (3 * 4)`:

1. The service receives the operation with a nested multiplication
2. It delegates the right operand `(3 * 4)` to the Multiplication Service
3. It adds the left operand `5` to the result `12`
4. Returns the final result `17` along with the complete event log
