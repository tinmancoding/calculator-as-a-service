# Multiplication Service

The Multiplication Service is a Go microservice that performs multiplication operations as part of the Calculator-as-a-Service (CaaS) architecture.

## Overview

- **Language:** Go 1.21
- **Port:** 8084
- **Operator:** `*` (multiplication)

## API Endpoints

### POST /execute

Execute a multiplication operation.

**Request:**
```json
{
  "operation": {
    "type": "operation",
    "operator": "*",
    "left": 5,
    "right": 3
  }
}
```

**Response:**
```json
{
  "result": 15,
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
      "service": "multiplication-service",
      "operation": "*",
      "operands": {
        "left": 5,
        "right": 3
      },
      "result": 15,
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
  "service": "multiplication-service"
}
```

### GET /ready

Readiness probe endpoint.

**Response:**
```json
{
  "status": "ready",
  "service": "multiplication-service"
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Name of this service | `multiplication-service` |
| `PORT` | Port to listen on | `8084` |
| `ADDITION_SERVICE_URL` | URL of addition service | `http://addition-service:8082` |
| `SUBTRACTION_SERVICE_URL` | URL of subtraction service | `http://subtraction-service:8083` |
| `DIVISION_SERVICE_URL` | URL of division service | `http://division-service:8086` |
| `MULTIPLICATION_SERVICE_URL` | URL of multiplication service | `http://multiplication-service:8084` |

## Running Locally

### Development Mode

```bash
# Run the service
go run main.go
```

### Production Mode

```bash
# Build the binary
go build -o multiplication-service main.go

# Run the binary
./multiplication-service
```

### Docker

```bash
# Build the image
docker build -t calculator/multiplication-service:latest .

# Run the container
docker run -p 8084:8084 calculator/multiplication-service:latest
```

## Kubernetes Deployment

```bash
# Apply the deployment and service
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## Delegation

The Multiplication Service can handle nested operations by delegating to other operation services. For example, when processing `(3 + 5) * (10 - 2)`:

1. The service receives the operation with nested addition and subtraction
2. It delegates the left operand `(3 + 5)` to the Addition Service
3. It delegates the right operand `(10 - 2)` to the Subtraction Service
4. It multiplies the results: `8 * 8 = 64`
5. Returns the final result along with the complete event log
