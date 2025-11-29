# Calculator as a Service (CaaS) - System Design

## Overview
A humorous yet educational microservices architecture that implements a basic arithmetic calculator where each operation is a separate microservice. Perfect for learning Kubernetes/OpenShift deployment strategies.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                          External Client                             │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 │ POST /calculate {"expression": "2+3*4"}
                                 ▼
                    ┌────────────────────────┐
                    │   Gateway Service      │
                    │   (Main Entry Point)   │
                    │   Port: 8080           │
                    └───────────┬────────────┘
                                │
                                │ POST /parse {"expression": "2+3*4"}
                                ▼
                    ┌────────────────────────┐
                    │   Parser Service       │
                    │   (Expression Parser)  │
                    │   Port: 8081           │
                    └───────────┬────────────┘
                                │
                                │ Returns: JSON AST
                                │ {"op": "+", "left": 2, 
                                │  "right": {"op": "*", "left": 3, "right": 4}}
                                ▼
                    ┌────────────────────────┐
                    │   Gateway Service      │
                    │  (Evaluates AST)       │
                    └───────────┬────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
    │  Addition    │  │Multiplication│  │  Division    │
    │  Service     │  │   Service    │  │   Service    │
    │  Port: 8082  │  │  Port: 8084  │  │  Port: 8086  │
    │  (Python)    │  │  (Go)        │  │  (Java)      │
    └──────────────┘  └──────────────┘  └──────────────┘
                │               │               │
                └───────────────┼───────────────┘
                                │
                    ┌──────────────┐
                    │ Subtraction  │
                    │   Service    │
                    │  Port: 8083  │
                    │ (TypeScript) │
                    └──────────────┘
```

---

## 1. Core Components

### 1.1 Gateway Service
**Responsibility:** Main entry point, orchestrates calculation requests

**Technology:** Node.js/Express or Go (recommend Go for simplicity)

**Endpoints:**
- `POST /calculate` - Main calculation endpoint
  - Request: `{"expression": "2 + 3 * 4"}`
  - Response: `{"result": 14, "eventLog": [...]}`

**Workflow:**
1. Receive expression string
2. Call Parser Service to convert to AST
3. Recursively evaluate AST by calling appropriate operation services
4. Collect event logs from all service calls
5. Return final result with complete event log

**Environment Variables:**
```bash
PARSER_SERVICE_URL=http://parser-service:8081
ADDITION_SERVICE_URL=http://addition-service:8082
SUBTRACTION_SERVICE_URL=http://subtraction-service:8083
MULTIPLICATION_SERVICE_URL=http://multiplication-service:8084
DIVISION_SERVICE_URL=http://division-service:8086
```

---

### 1.2 Parser Service
**Responsibility:** Convert string expressions to JSON Abstract Syntax Tree

**Technology:** Python (Flask with custom recursive descent parser)

**Endpoints:**
- `POST /parse`
  - Request: `{"expression": "2 + 3 * 4"}`
  - Response: `{"ast": {...}, "eventLog": [...]}`
- `GET /health` - Liveness probe
- `GET /ready` - Readiness probe

**AST Structure:**
```json
{
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
}
```

**Features:**
- Operator precedence (*, / before +, -)
- Parentheses support
- Decimal numbers support (using pattern: `\d+\.\d+|\d+`)
- Error handling for invalid expressions
- Input validation (max expression length: 1000 characters)
- Security: Internal errors logged server-side, generic errors returned to clients

**Environment Variables:**
```bash
SERVICE_NAME=parser-service
PORT=8081
MAX_EXPRESSION_LENGTH=1000  # Maximum expression length to prevent abuse
```

**Grammar:**
```
expression  : term (('+' | '-') term)*
term        : factor (('*' | '/') factor)*
factor      : NUMBER | '(' expression ')'
NUMBER      : \d+\.\d+ | \d+  # Decimal or integer
```

---

### 1.3 Operation Services

Each operation service implements the **same API contract** but in different languages.

#### Common API Contract

**Endpoint:** `POST /execute`

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

Or for nested operations:
```json
{
  "operation": {
    "type": "operation",
    "operator": "+",
    "left": 5,
    "right": {
      "type": "operation",
      "operator": "*",
      "left": 3,
      "right": 4
    }
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
      }
    }
  ]
}
```

**Response for Delegated Operations:**
```json
{
  "result": 17,
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:31.087Z",
      "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
      "service": "multiplication-service",
      "operation": "*",
      "operands": {
        "left": 3,
        "right": 4
      },
      "result": 12,
      "delegations": {
        "left": null,
        "right": null
      }
    },
    {
      "timestamp": "2024-11-29T10:15:32.145Z",
      "hostname": "addition-service-7d9f8b6c5-xk2m9",
      "service": "addition-service",
      "operation": "+",
      "operands": {
        "left": 5,
        "right": 12
      },
      "result": 17,
      "delegations": {
        "left": null,
        "right": {
          "service": "multiplication-service",
          "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
          "operation": "*",
          "result": 12
        }
      }
    }
  ]
}
```

#### 1.3.1 Addition Service (Python/Flask)
**Path:** `/services/addition-service`
**Port:** 8082
**Environment Variables:**
```bash
SERVICE_NAME=addition-service
MULTIPLICATION_SERVICE_URL=http://multiplication-service:8084
DIVISION_SERVICE_URL=http://division-service:8086
SUBTRACTION_SERVICE_URL=http://subtraction-service:8083
```

#### 1.3.2 Subtraction Service (TypeScript/Express)
**Path:** `/services/subtraction-service`
**Port:** 8083
**Environment Variables:** (similar to addition)

#### 1.3.3 Multiplication Service (Go)
**Path:** `/services/multiplication-service`
**Port:** 8084
**Environment Variables:** (similar to addition)

#### 1.3.4 Division Service (Java Spring Boot)
**Path:** `/services/division-service`
**Port:** 8086
**Additional Feature:** Division by zero error handling

---

## 2. Data Models

### 2.1 AST Node Types

```json
// Number Node
{
  "type": "number",
  "value": 42
}

// Operation Node
{
  "type": "operation",
  "operator": "+|-|*|/",
  "left": <ASTNode>,
  "right": <ASTNode>
}
```

### 2.2 Event Log Entry

```json
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
```

**Delegation Object Structure** (when an operand is delegated):
```json
{
  "service": "multiplication-service",
  "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
  "operation": "*",
  "result": 12
}
```

**Complete Example with Delegations:**
```json
{
  "timestamp": "2024-11-29T10:15:32.145Z",
  "hostname": "addition-service-7d9f8b6c5-xk2m9",
  "service": "addition-service",
  "operation": "+",
  "operands": {
    "left": 5,
    "right": 12
  },
  "result": 17,
  "delegations": {
    "left": null,
    "right": {
      "service": "multiplication-service",
      "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
      "operation": "*",
      "result": 12
    }
  },
  "duration": 8
}
```

### 2.3 API Response Format

```json
{
  "result": "final calculated value",
  "expression": "original expression string",
  "eventLog": [
    // Array of EventLogEntry objects in chronological order
  ],
  "metadata": {
    "totalServices": "number of services involved",
    "totalDuration": "total processing time"
  }
}
```

---

## 3. Service Implementation Details

### 3.1 Service Delegation Logic

Each operation service must:

1. **Check operand types**
   - If both operands are numbers → perform operation directly
   - If one/both operands are operation nodes → delegate

2. **Delegate to appropriate service**
   ```
   function evaluate(operand, side):  // side is "left" or "right"
     if operand.type == "number":
       return {
         value: operand.value,
         delegation: null,
         eventLogs: []
       }
     else if operand.type == "operation":
       response = delegate to service for operand.operator
       return {
         value: response.result,
         delegation: {
           service: response.service,
           hostname: response.hostname,
           operation: operand.operator,
           result: response.result
         },
         eventLogs: response.eventLog
       }
   ```

3. **Collect event logs and build delegation info**
   ```
   leftEval = evaluate(operation.left, "left")
   rightEval = evaluate(operation.right, "right")
   
   result = performOperation(leftEval.value, rightEval.value)
   
   myEvent = {
     timestamp: getCurrentTimestamp(),
     hostname: getHostname(),
     service: getServiceName(),
     operation: operation.operator,
     operands: {
       left: leftEval.value,
       right: rightEval.value
     },
     result: result,
     delegations: {
       left: leftEval.delegation,
       right: rightEval.delegation
     },
     duration: getDuration()
   }
   
   allEvents = [...leftEval.eventLogs, ...rightEval.eventLogs, myEvent]
   return { result, eventLog: allEvents }
   ```

4. **Hostname Retrieval**
   Each service should get its hostname from the environment:
   ```
   hostname = os.getenv('HOSTNAME') or socket.gethostname()
   ```
   In Kubernetes, the HOSTNAME environment variable is automatically set to the pod name.

### 3.2 Example: Addition Service Processing "5 + (3 * 4)"

```
1. Addition Service (pod: addition-service-7d9f8b6c5-xk2m9) receives:
   {op: "+", left: 5, right: {op: "*", left: 3, right: 4}}

2. Evaluates left operand:
   - Type is "number" with value 5
   - No delegation needed
   - leftEval = {value: 5, delegation: null, eventLogs: []}

3. Evaluates right operand:
   - Type is "operation" with operator "*"
   - Calls Multiplication Service: POST http://multiplication-service:8084/execute
   - Multiplication Service (pod: multiplication-service-5b8c9d7f6-p4q2k) returns:
     {
       result: 12,
       eventLog: [{
         timestamp: "2024-11-29T10:15:31.087Z",
         hostname: "multiplication-service-5b8c9d7f6-p4q2k",
         service: "multiplication-service",
         operation: "*",
         operands: {left: 3, right: 4},
         result: 12,
         delegations: {left: null, right: null}
       }]
     }
   - rightEval = {
       value: 12,
       delegation: {
         service: "multiplication-service",
         hostname: "multiplication-service-5b8c9d7f6-p4q2k",
         operation: "*",
         result: 12
       },
       eventLogs: [<multiplication event>]
     }

4. Addition Service performs: 5 + 12 = 17

5. Addition Service builds its event:
   {
     timestamp: "2024-11-29T10:15:32.145Z",
     hostname: "addition-service-7d9f8b6c5-xk2m9",
     service: "addition-service",
     operation: "+",
     operands: {left: 5, right: 12},
     result: 17,
     delegations: {
       left: null,
       right: {
         service: "multiplication-service",
         hostname: "multiplication-service-5b8c9d7f6-p4q2k",
         operation: "*",
         result: 12
       }
     }
   }

6. Addition Service returns:
   {
     result: 17,
     eventLog: [
       <multiplication event>,
       <addition event>
     ]
   }
```

**Key Points:**
- The `delegations.left` is `null` because left operand (5) was a direct number
- The `delegations.right` shows we delegated to multiplication-service
- The hostname shows which specific pod handled each operation
- Event logs are accumulated in chronological order

---

## 4. Kubernetes/OpenShift Deployment

### 4.1 Deployment Strategy

Each microservice gets:
- **Deployment** (with configurable replicas for scaling demos)
- **Service** (ClusterIP for internal communication)
- **ConfigMap** (for service URLs)
- **Optional: HorizontalPodAutoscaler** (for scaling demos)

### 4.2 Service Discovery

Use Kubernetes DNS:
```
http://addition-service.default.svc.cluster.local:8082
```

Or simplified (within same namespace):
```
http://addition-service:8082
```

### 4.3 Example Deployment Structure

```yaml
# addition-service/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: addition-service
  labels:
    app: calculator
    service: addition
spec:
  replicas: 2
  selector:
    matchLabels:
      app: calculator
      service: addition
  template:
    metadata:
      labels:
        app: calculator
        service: addition
    spec:
      containers:
      - name: addition
        image: calculator/addition-service:latest
        imagePullPolicy: Always  # Use Always with :latest tag for consistency
        ports:
        - containerPort: 8082
        env:
        - name: SERVICE_NAME
          value: "addition-service"
        - name: MULTIPLICATION_SERVICE_URL
          valueFrom:
            configMapKeyRef:
              name: calculator-config
              key: multiplication_service_url
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: addition-service
spec:
  selector:
    app: calculator
    service: addition
  ports:
  - port: 8082
    targetPort: 8082
  type: ClusterIP
```

### 4.4 Gateway Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: calculator-gateway
spec:
  rules:
  - host: calculator.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gateway-service
            port:
              number: 8080
```

---

## 5. Project Structure

```
calculator-as-a-service/
├── README.md
├── docs/
│   ├── ARCHITECTURE.md
│   └── API.md
├── k8s/
│   ├── namespace.yaml
│   ├── configmap.yaml
│   └── ingress.yaml
├── services/
│   ├── gateway/
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   ├── k8s/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── README.md
│   ├── parser/
│   │   ├── Dockerfile
│   │   ├── requirements.txt
│   │   ├── app.py
│   │   ├── parser.py
│   │   ├── k8s/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── README.md
│   ├── addition-service/
│   │   ├── Dockerfile
│   │   ├── requirements.txt
│   │   ├── app.py
│   │   ├── k8s/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── README.md
│   ├── subtraction-service/
│   │   ├── Dockerfile
│   │   ├── package.json
│   │   ├── src/
│   │   │   └── index.ts
│   │   ├── k8s/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── README.md
│   ├── multiplication-service/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── main.go
│   │   ├── k8s/
│   │   │   ├── deployment.yaml
│   │   │   └── service.yaml
│   │   └── README.md
│   └── division-service/
│       ├── Dockerfile
│       ├── pom.xml
│       ├── src/main/java/com/calculator/division/
│       │   ├── DivisionApplication.java
│       │   └── DivisionController.java
│       ├── k8s/
│       │   ├── deployment.yaml
│       │   └── service.yaml
│       └── README.md
├── scripts/
│   ├── build-all.sh
│   ├── deploy-all.sh
│   └── test-calculator.sh
└── examples/
    └── sample-requests.http
```

---

## 6. Learning Opportunities

This architecture demonstrates:

### 6.1 Kubernetes/OpenShift Concepts
- **Service Discovery:** DNS-based service-to-service communication
- **ConfigMaps:** Environment configuration management
- **Deployments:** Rolling updates and scaling
- **Services:** ClusterIP, NodePort, LoadBalancer
- **Ingress:** External traffic routing
- **Health Checks:** Liveness and readiness probes
- **Resource Management:** Requests and limits

### 6.2 Deployment Strategies
- **Blue/Green Deployment:** Deploy new version alongside old
- **Canary Deployment:** Gradually shift traffic to new version
- **Rolling Update:** Default Kubernetes update strategy
- **Horizontal Scaling:** Scale services independently

### 6.3 Observability
- **Distributed Tracing:** Event logs show service interactions
- **Logging:** Structured logs from each service
- **Metrics:** Prometheus metrics (requests, latency, errors)

### 6.4 Resilience Patterns
- **Retry Logic:** Handle transient failures
- **Circuit Breaker:** Prevent cascading failures
- **Timeout Handling:** Graceful degradation
- **Health Checks:** Automatic pod replacement

---

## 7. Example Scenarios

### 7.1 Simple Addition: "5 + 3"

**Request:**
```bash
curl -X POST http://calculator.example.com/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "5 + 3"}'
```

**Event Log:**
```json
{
  "result": 8,
  "expression": "5 + 3",
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:15:30.123Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "5 + 3",
      "result": "AST generated",
      "duration": 5
    },
    {
      "timestamp": "2024-11-29T10:15:30.150Z",
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
  ],
  "metadata": {
    "totalServices": 2,
    "totalDuration": 7
  }
}
```

### 7.2 Complex Expression: "10 + 5 * 2"

**Event Log:**
```json
{
  "result": 20,
  "expression": "10 + 5 * 2",
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:16:00.123Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "10 + 5 * 2",
      "result": "AST generated"
    },
    {
      "timestamp": "2024-11-29T10:16:00.155Z",
      "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
      "service": "multiplication-service",
      "operation": "*",
      "operands": {
        "left": 5,
        "right": 2
      },
      "result": 10,
      "delegations": {
        "left": null,
        "right": null
      }
    },
    {
      "timestamp": "2024-11-29T10:16:00.180Z",
      "hostname": "addition-service-7d9f8b6c5-xk2m9",
      "service": "addition-service",
      "operation": "+",
      "operands": {
        "left": 10,
        "right": 10
      },
      "result": 20,
      "delegations": {
        "left": null,
        "right": {
          "service": "multiplication-service",
          "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
          "operation": "*",
          "result": 10
        }
      }
    }
  ]
}
```

### 7.3 Nested Expression: "(3 + 5) * (10 - 2)"

**Event Log:**
```json
{
  "result": 64,
  "expression": "(3 + 5) * (10 - 2)",
  "eventLog": [
    {
      "timestamp": "2024-11-29T10:17:00.100Z",
      "hostname": "parser-service-6c8d9f7b5-m3n4p",
      "service": "parser-service",
      "operation": "parse",
      "input": "(3 + 5) * (10 - 2)",
      "result": "AST generated"
    },
    {
      "timestamp": "2024-11-29T10:17:00.125Z",
      "hostname": "addition-service-7d9f8b6c5-xk2m9",
      "service": "addition-service",
      "operation": "+",
      "operands": {
        "left": 3,
        "right": 5
      },
      "result": 8,
      "delegations": {
        "left": null,
        "right": null
      }
    },
    {
      "timestamp": "2024-11-29T10:17:00.145Z",
      "hostname": "subtraction-service-8e7f6d5c4-q9r8s",
      "service": "subtraction-service",
      "operation": "-",
      "operands": {
        "left": 10,
        "right": 2
      },
      "result": 8,
      "delegations": {
        "left": null,
        "right": null
      }
    },
    {
      "timestamp": "2024-11-29T10:17:00.170Z",
      "hostname": "multiplication-service-5b8c9d7f6-p4q2k",
      "service": "multiplication-service",
      "operation": "*",
      "operands": {
        "left": 8,
        "right": 8
      },
      "result": 64,
      "delegations": {
        "left": {
          "service": "addition-service",
          "hostname": "addition-service-7d9f8b6c5-xk2m9",
          "operation": "+",
          "result": 8
        },
        "right": {
          "service": "subtraction-service",
          "hostname": "subtraction-service-8e7f6d5c4-q9r8s",
          "operation": "-",
          "result": 8
        }
      }
    }
  ]
}
```

**Note:** This example clearly shows that the multiplication service delegated:
- **Left operand** to the addition-service (3 + 5 = 8)
- **Right operand** to the subtraction-service (10 - 2 = 8)
- Then performed its own operation: 8 * 8 = 64

---

## 8. Testing Strategy

### 8.1 Unit Tests
Each service should have unit tests for:
- Basic operation logic
- AST parsing (parser service)
- Error handling

### 8.2 Integration Tests
- Service-to-service communication
- Event log aggregation
- End-to-end expression evaluation

### 8.3 Load Testing
```bash
# Using k6 or Apache Bench
k6 run --vus 10 --duration 30s load-test.js
```

### 8.4 Chaos Engineering
- Kill random pods during calculation
- Introduce network latency
- Test circuit breaker behavior

---

## 9. Monitoring & Observability

### 9.1 Metrics to Track
- **Request rate:** Requests per second per service
- **Latency:** P50, P95, P99 response times
- **Error rate:** 4xx and 5xx responses
- **Service dependencies:** Call graph visualization

### 9.2 Prometheus Metrics
Each service exposes `/metrics`:
```
calculator_requests_total{service="addition",operation="+"}
calculator_request_duration_seconds{service="addition"}
calculator_errors_total{service="addition",error_type="delegation_failed"}
```

### 9.3 Grafana Dashboard
Create dashboard showing:
- Request flow through services
- Service health status
- Event log visualization
- Expression evaluation tree

---

## 10. Advanced Features (Future Enhancements)

### 10.1 Caching Layer
Add Redis to cache frequently calculated expressions

### 10.2 Rate Limiting
Implement per-service rate limiting

### 10.3 Authentication
Add JWT-based API authentication

### 10.4 Asynchronous Processing
Use message queue (RabbitMQ/Kafka) for complex calculations

### 10.5 Expression History
Store calculation history in database

### 10.6 WebSocket Support
Real-time calculation updates for long-running operations

### 10.7 Additional Operations
- Power (exponentiation)
- Modulo
- Square root
- Factorial (demonstrating recursion)

---

## 11. Workshop Exercises

### Exercise 1: Deploy Basic Services
Deploy gateway, parser, and one operation service

### Exercise 2: Add More Operations
Deploy remaining operation services

### Exercise 3: Scale Horizontally
Scale multiplication service to 5 replicas

### Exercise 4: Rolling Update
Update addition service with zero downtime

### Exercise 5: Blue/Green Deployment
Deploy new parser version using blue/green strategy

### Exercise 6: Implement Circuit Breaker
Add circuit breaker to prevent cascading failures

### Exercise 7: Add Monitoring
Set up Prometheus and Grafana

### Exercise 8: Chaos Testing
Introduce random pod failures and observe recovery

---

## 12. Common Pitfalls & Solutions

### 12.1 Circular Dependencies
**Problem:** Services calling each other in loops
**Solution:** AST evaluation order ensures no cycles

### 12.2 Event Log Explosion
**Problem:** Logs grow too large for complex expressions
**Solution:** Implement log size limits and sampling

### 12.3 Service Discovery Failures
**Problem:** Services can't find each other
**Solution:** Use proper DNS names and health checks

### 12.4 Resource Exhaustion
**Problem:** Too many service replicas consuming resources
**Solution:** Set proper resource limits and HPA thresholds

### 12.5 Security Vulnerabilities
**Problem:** Input validation missing, internal errors exposed to clients
**Solution:**
- Implement input validation (e.g., MAX_EXPRESSION_LENGTH for parser)
- Log detailed errors server-side for debugging
- Return generic error messages to clients to avoid information leakage
- Use proper error boundaries in all services

### 12.6 Image Deployment Inconsistencies
**Problem:** Using `:latest` tag with `imagePullPolicy: IfNotPresent` causes inconsistent deployments
**Solution:**
- Use `imagePullPolicy: Always` with `:latest` tags
- Or use specific version tags (e.g., `v1.2.3`) with `IfNotPresent`

---

## Conclusion

This Calculator-as-a-Service architecture provides a fun, hands-on way to learn microservices, Kubernetes, and distributed systems concepts. The deliberately over-engineered design makes it perfect for teaching:

- Service decomposition
- Inter-service communication
- Deployment strategies
- Observability and monitoring
- Resilience patterns
- Container orchestration

The event logging system is especially valuable for understanding how distributed systems work, as it provides complete visibility into the service interaction flow.
