# Division Service

The Division Service is a Java/Spring Boot microservice that performs division operations as part of the Calculator-as-a-Service (CaaS) architecture.

## Overview

- **Language:** Java 17 with Spring Boot 3.2
- **Port:** 8086
- **Operator:** `/` (division)

## API Endpoints

### POST /execute

Execute a division operation.

**Request:**
```json
{
  "operation": {
    "type": "operation",
    "operator": "/",
    "left": 10,
    "right": 2
  }
}
```

**Response:**
```json
{
  "result": 5.0,
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "division-service-7d9f8b6c5-xk2m9",
      "service": "division-service",
      "operation": "/",
      "operands": {
        "left": 10.0,
        "right": 2.0
      },
      "result": 5.0,
      "delegations": {
        "left": null,
        "right": null
      },
      "duration": 2
    }
  ]
}
```

### Division by Zero Error

When attempting to divide by zero, the service returns an error:

**Request:**
```json
{
  "operation": {
    "type": "operation",
    "operator": "/",
    "left": 10,
    "right": 0
  }
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Division by zero"
}
```

### GET /health

Liveness probe endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "division-service"
}
```

### GET /ready

Readiness probe endpoint.

**Response:**
```json
{
  "status": "ready",
  "service": "division-service"
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Name of this service | `division-service` |
| `ADDITION_SERVICE_URL` | URL of addition service | `http://addition-service:8082` |
| `SUBTRACTION_SERVICE_URL` | URL of subtraction service | `http://subtraction-service:8083` |
| `MULTIPLICATION_SERVICE_URL` | URL of multiplication service | `http://multiplication-service:8084` |
| `DIVISION_SERVICE_URL` | URL of division service | `http://division-service:8086` |

## Running Locally

### Prerequisites

- Java 17 or higher
- Maven 3.6 or higher

### Development Mode

```bash
# Build and run the service
./mvnw spring-boot:run
```

### Production Mode

```bash
# Build the application
./mvnw package

# Run the jar
java -jar target/division-service-1.0.0.jar
```

### Docker

```bash
# Build the image
docker build -t calculator/division-service:latest .

# Run the container
docker run -p 8086:8086 calculator/division-service:latest
```

## Kubernetes Deployment

```bash
# Apply the deployment and service
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## Delegation

The Division Service can handle nested operations by delegating to other operation services. For example, when processing `(10 + 2) / (3 - 1)`:

1. The service receives the operation with nested addition and subtraction
2. It delegates the left operand `(10 + 2)` to the Addition Service
3. It delegates the right operand `(3 - 1)` to the Subtraction Service
4. It divides the results: `12 / 2 = 6`
5. Returns the final result along with the complete event log
