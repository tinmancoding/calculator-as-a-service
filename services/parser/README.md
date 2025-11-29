# Parser Service

The Parser Service is a Python-based microservice that converts arithmetic expressions into Abstract Syntax Trees (AST) for the Calculator-as-a-Service platform.

## Overview

This service implements a recursive descent parser that handles:
- Basic arithmetic operations: `+`, `-`, `*`, `/`
- Operator precedence (multiplication and division before addition and subtraction)
- Parentheses for grouping
- Decimal numbers
- Comprehensive error handling

## Technology Stack

- **Language**: Python 3.11
- **Framework**: Flask
- **WSGI Server**: Gunicorn
- **Port**: 8081

## API Endpoints

### POST /parse

Converts an arithmetic expression string to an AST.

**Request:**
```json
{
  "expression": "2 + 3 * 4"
}
```

**Response (Success):**
```json
{
  "ast": {
    "type": "operation",
    "operator": "+",
    "left": {
      "type": "number",
      "value": 2
    },
    "right": {
      "type": "operation",
      "operator": "*",
      "left": {
        "type": "number",
        "value": 3
      },
      "right": {
        "type": "number",
        "value": 4
      }
    }
  },
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "2 + 3 * 4",
      "result": "AST generated",
      "duration": 5
    }
  ]
}
```

**Response (Error):**
```json
{
  "error": "Unexpected token: )",
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "2 + )",
      "result": "Parse error: Unexpected token: )",
      "duration": 2
    }
  ]
}
```

### GET /health

Health check endpoint for Kubernetes liveness probe.

**Response:**
```json
{
  "status": "healthy",
  "service": "parser-service",
  "hostname": "parser-service-6c8d9f7b5-m3n4p"
}
```

### GET /ready

Readiness check endpoint for Kubernetes readiness probe.

**Response:**
```json
{
  "status": "ready",
  "service": "parser-service",
  "hostname": "parser-service-6c8d9f7b5-m3n4p"
}
```

### GET /

Service information endpoint.

**Response:**
```json
{
  "service": "parser-service",
  "version": "1.0.0",
  "hostname": "parser-service-6c8d9f7b5-m3n4p",
  "endpoints": {
    "parse": "POST /parse",
    "health": "GET /health",
    "ready": "GET /ready"
  }
}
```

## AST Structure

The parser generates an AST with two node types:

### Number Node
```json
{
  "type": "number",
  "value": 42
}
```

### Operation Node
```json
{
  "type": "operation",
  "operator": "+|-|*|/",
  "left": <ASTNode>,
  "right": <ASTNode>
}
```

## Grammar

The parser implements the following grammar:

```
expression  : term (('+' | '-') term)*
term        : factor (('*' | '/') factor)*
factor      : NUMBER | '(' expression ')'
NUMBER      : [0-9]+('.'[0-9]+)?
```

This grammar ensures correct operator precedence:
- Multiplication and division have higher precedence than addition and subtraction
- Parentheses can override default precedence

## Example Expressions

### Simple Addition: `5 + 3`
```json
{
  "type": "operation",
  "operator": "+",
  "left": {"type": "number", "value": 5},
  "right": {"type": "number", "value": 3}
}
```

### Operator Precedence: `10 + 5 * 2`
Multiplication evaluated first: `10 + (5 * 2)`
```json
{
  "type": "operation",
  "operator": "+",
  "left": {"type": "number", "value": 10},
  "right": {
    "type": "operation",
    "operator": "*",
    "left": {"type": "number", "value": 5},
    "right": {"type": "number", "value": 2}
  }
}
```

### Parentheses Override: `(3 + 5) * 2`
```json
{
  "type": "operation",
  "operator": "*",
  "left": {
    "type": "operation",
    "operator": "+",
    "left": {"type": "number", "value": 3},
    "right": {"type": "number", "value": 5}
  },
  "right": {"type": "number", "value": 2}
}
```

### Decimal Numbers: `10.5 + 2.3`
```json
{
  "type": "operation",
  "operator": "+",
  "left": {"type": "number", "value": 10.5},
  "right": {"type": "number", "value": 2.3}
}
```

### Nested Parentheses: `((1 + 2) * 3) / 4`
```json
{
  "type": "operation",
  "operator": "/",
  "left": {
    "type": "operation",
    "operator": "*",
    "left": {
      "type": "operation",
      "operator": "+",
      "left": {"type": "number", "value": 1},
      "right": {"type": "number", "value": 2}
    },
    "right": {"type": "number", "value": 3}
  },
  "right": {"type": "number", "value": 4}
}
```

## Local Development

### Prerequisites
- Python 3.11 or higher
- pip

### Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Run the service:
```bash
python app.py
```

The service will start on `http://localhost:8081`

### Testing

Test the parser with curl:

```bash
# Simple expression
curl -X POST http://localhost:8081/parse \
  -H "Content-Type: application/json" \
  -d '{"expression": "2 + 3"}'

# Complex expression with precedence
curl -X POST http://localhost:8081/parse \
  -H "Content-Type: application/json" \
  -d '{"expression": "10 + 5 * 2"}'

# Parentheses
curl -X POST http://localhost:8081/parse \
  -H "Content-Type: application/json" \
  -d '{"expression": "(3 + 5) * (10 - 2)"}'

# Health check
curl http://localhost:8081/health
```

## Docker

### Build Image

```bash
docker build -t calculator/parser-service:latest .
```

### Run Container

```bash
docker run -p 8081:8081 \
  -e SERVICE_NAME=parser-service \
  -e PORT=8081 \
  calculator/parser-service:latest
```

## Kubernetes Deployment

### Deploy

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -l service=parser

# Check service
kubectl get svc parser-service

# View logs
kubectl logs -l service=parser -f
```

### Test in Cluster

```bash
# Port forward
kubectl port-forward svc/parser-service 8081:8081

# Test
curl -X POST http://localhost:8081/parse \
  -H "Content-Type: application/json" \
  -d '{"expression": "2 + 3 * 4"}'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `parser-service` | Name of the service for logging |
| `PORT` | `8081` | Port the service listens on |
| `HOSTNAME` | System hostname | Pod name in Kubernetes |
| `MAX_EXPRESSION_LENGTH` | `1000` | Maximum allowed length of the input expression |

## Error Handling

The parser handles various error conditions:

| Error | Description |
|-------|-------------|
| Empty expression | Expression is empty or whitespace only |
| Invalid characters | Expression contains unsupported characters |
| Unexpected token | Token appears in wrong position |
| Missing closing parenthesis | Unmatched opening parenthesis |
| Unexpected end of expression | Expression ends prematurely |

## Performance Considerations

- **Parsing complexity**: O(n) where n is the length of the expression
- **Memory usage**: O(n) for the token list and AST
- **Workers**: Configured with 2 Gunicorn workers by default
- **Timeout**: 30 seconds for request processing

## Monitoring

### Metrics to Monitor
- Request rate (requests/second)
- Parse latency (p50, p95, p99)
- Error rate (4xx, 5xx responses)
- Pod restarts
- Memory usage

### Health Checks
- **Liveness**: `/health` endpoint (checks if service is running)
- **Readiness**: `/ready` endpoint (checks if service can accept traffic)

## Integration

The Parser Service is designed to be called by the Gateway Service:

```
Client → Gateway Service → Parser Service
                        ↓
                    Returns AST
                        ↓
          Gateway evaluates AST using operation services
```

## Architecture Notes

This service is intentionally kept stateless and simple:
- No database dependencies
- No caching (parsing is fast enough)
- No external service calls
- Pure computation service

This makes it ideal for horizontal scaling and demonstrates microservice best practices.
