# calculator-as-a-service
Calculator as a Service (CaaS) is an example microservice architecture used on my Kubernetes/OpenShift/Podman workshop

## Quick Start

### Deploy with Helm (Recommended)

The easiest way to deploy all services to Kubernetes or OpenShift:

```bash
# Clone the repository first
git clone https://github.com/tinmancoding/calculator-as-a-service.git
cd calculator-as-a-service

# Standard Kubernetes deployment (using pre-built images)
helm install calculator ./helm/calculator --namespace calculator --create-namespace

# OpenShift with source-to-image builds (builds from GitHub)
# Note: On OpenShift sandbox, use your existing project instead of --create-namespace
helm install calculator ./helm/calculator \
  --set openshift.build.enabled=true \
  --set openshift.build.sourceRepo=https://github.com/tinmancoding/calculator-as-a-service.git
```

See the [Helm Quick Start Guide](./helm/QUICKSTART.md) for detailed instructions.

### Deploy with Docker Compose

```bash
docker-compose up -d
```

## Architecture

This project demonstrates a distributed microservices calculator with 6 services:
- **Gateway Service** (Go) - Entry point and request routing
- **Parser Service** (Python) - Expression parsing  
- **Addition Service** (Python) - Addition operations
- **Subtraction Service** (TypeScript/Node.js) - Subtraction operations
- **Multiplication Service** (Go) - Multiplication operations
- **Division Service** (Java/Spring Boot) - Division operations

## Deployment Options

1. **Helm Chart** - Automated Kubernetes/OpenShift deployment (recommended)
   - [Helm Chart Documentation](./helm/calculator/README.md)
   - [Quick Start Guide](./helm/QUICKSTART.md)

2. **Docker Compose** - Local development
   - See `docker-compose.yaml`

3. **Manual Kubernetes** - Individual service deployment
   - Each service has k8s manifests in `services/*/k8s/`

## Documentation

- [System Design](./SYSTEM-DESIGN.md) - Detailed architecture and design decisions
- [Helm Chart README](./helm/calculator/README.md) - Helm deployment guide
- [Quick Start](./helm/QUICKSTART.md) - Get started in 5 minutes

