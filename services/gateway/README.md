# Gateway Service

The Gateway Service is the main entry point for the Calculator-as-a-Service system. It orchestrates calculation requests by coordinating with the Parser Service and various Operation Services.

## Technology

- **Language:** Go 1.21
- **Port:** 8080

## Endpoints

### POST /calculate

Main calculation endpoint that accepts arithmetic expressions and returns the computed result.

**Request:**
```json
{
  "expression": "2 + 3 * 4"
}
```

**Response:**
```json
{
  "result": 14,
  "expression": "2 + 3 * 4",
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "2 + 3 * 4",
      "result": "AST generated",
      "duration": 5
    },
    ...
  ],
  "metadata": {
    "totalServices": 3,
    "totalDuration": 15
  }
}
```

### GET /health

Liveness probe endpoint for Kubernetes.

**Response:**
```json
{
  "status": "healthy",
  "service": "gateway-service",
  "hostname": "gateway-service-7d9f8b6c5-xk2m9"
}
```

### GET /ready

Readiness probe endpoint for Kubernetes.

**Response:**
```json
{
  "status": "ready",
  "service": "gateway-service",
  "hostname": "gateway-service-7d9f8b6c5-xk2m9"
}
```

### GET /

Service information endpoint.

**Response:**
```json
{
  "service": "gateway-service",
  "version": "1.0.0",
  "hostname": "gateway-service-7d9f8b6c5-xk2m9",
  "endpoints": {
    "calculate": "POST /calculate",
    "health": "GET /health",
    "ready": "GET /ready"
  }
}
```

## Workflow

1. Receive expression string via POST /calculate
2. Call Parser Service to convert expression to AST
3. Recursively evaluate AST by calling appropriate operation services
4. Collect event logs from all service calls
5. Return final result with complete event log and metadata

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Port to listen on |
| `SERVICE_NAME` | gateway-service | Service name for logging |
| `PARSER_SERVICE_URL` | http://parser-service:8081 | Parser service URL |
| `ADDITION_SERVICE_URL` | http://addition-service:8082 | Addition service URL |
| `SUBTRACTION_SERVICE_URL` | http://subtraction-service:8083 | Subtraction service URL |
| `MULTIPLICATION_SERVICE_URL` | http://multiplication-service:8084 | Multiplication service URL |
| `DIVISION_SERVICE_URL` | http://division-service:8086 | Division service URL |

## Building

### Local Build

```bash
go build -o gateway .
./gateway
```

### Docker Build

```bash
docker build -t calculator/gateway-service:latest .
docker run -p 8080:8080 calculator/gateway-service:latest
```

## Testing

```bash
# Health check
curl http://localhost:8080/health

# Calculate expression
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "5 + 3"}'
```

## Kubernetes Deployment

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```
