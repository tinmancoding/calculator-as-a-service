# Subtraction Service

The Subtraction Service is a TypeScript/Express microservice that performs subtraction operations as part of the Calculator-as-a-Service (CaaS) architecture.

## Overview

- **Language:** TypeScript with Express
- **Port:** 8083
- **Operator:** `-` (subtraction)

## API Endpoints

### POST /execute

Execute a subtraction operation.

**Request:**
```json
{
  "operation": {
    "type": "operation",
    "operator": "-",
    "left": 10,
    "right": 3
  }
}
```

**Response:**
```json
{
  "result": 7,
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "subtraction-service-8e7f6d5c4-q9r8s",
      "service": "subtraction-service",
      "operation": "-",
      "operands": {
        "left": 10,
        "right": 3
      },
      "result": 7,
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
  "service": "subtraction-service"
}
```

### GET /ready

Readiness probe endpoint.

**Response:**
```json
{
  "status": "ready",
  "service": "subtraction-service"
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Name of this service | `subtraction-service` |
| `PORT` | Port to listen on | `8083` |
| `ADDITION_SERVICE_URL` | URL of addition service | `http://addition-service:8082` |
| `MULTIPLICATION_SERVICE_URL` | URL of multiplication service | `http://multiplication-service:8084` |
| `DIVISION_SERVICE_URL` | URL of division service | `http://division-service:8086` |

## Running Locally

### Development Mode

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev
```

### Production Mode

```bash
# Install dependencies
npm install

# Build TypeScript
npm run build

# Run the compiled JavaScript
npm start
```

### Docker

```bash
# Build the image
docker build -t calculator/subtraction-service:latest .

# Run the container
docker run -p 8083:8083 calculator/subtraction-service:latest
```

## Kubernetes Deployment

```bash
# Apply the deployment and service
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## Delegation

The Subtraction Service can handle nested operations by delegating to other operation services. For example, when processing `10 - (3 * 2)`:

1. The service receives the operation with a nested multiplication
2. It delegates the right operand `(3 * 2)` to the Multiplication Service
3. It subtracts the result `6` from the left operand `10`
4. Returns the final result `4` along with the complete event log
