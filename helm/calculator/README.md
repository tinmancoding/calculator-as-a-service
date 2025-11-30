# Calculator as a Service - Helm Chart

This Helm chart deploys the Calculator as a Service distributed microservices application to Kubernetes or OpenShift.

## Overview

Calculator as a Service is a distributed calculator system consisting of 6 microservices:
- **Gateway Service**: Entry point and request routing
- **Parser Service**: Expression parsing
- **Addition Service**: Addition operations
- **Subtraction Service**: Subtraction operations
- **Multiplication Service**: Multiplication operations
- **Division Service**: Division operations

## Prerequisites

- Kubernetes 1.19+ or OpenShift 4.x+
- Helm 3.0+
- kubectl or oc CLI configured

## Installation

### Standard Kubernetes Deployment (using pre-built images)

```bash
# Install with default values
helm install calculator ./helm/calculator

# Install in a specific namespace
helm install calculator ./helm/calculator --namespace calculator --create-namespace

# Install with custom values
helm install calculator ./helm/calculator --values my-values.yaml
```

### OpenShift Deployment (with source-to-image builds)

```bash
# Enable OpenShift builds to build images from source
helm install calculator ./helm/calculator \
  --set openshift.build.enabled=true \
  --set openshift.build.sourceRepo=https://github.com/tinmancoding/calculator-as-a-service.git \
  --set openshift.build.sourceRef=main \
  --namespace calculator \
  --create-namespace

# Monitor the builds
oc get builds -w

# Once builds complete, the deployments will automatically use the built images
```

## Configuration

### Key Configuration Options

| Parameter | Description | Default |
|-----------|-------------|---------|
| `global.imagePullPolicy` | Image pull policy for all services | `Always` |
| `global.imageRegistry` | Docker registry for pre-built images | `tinmancoding` |
| `openshift.build.enabled` | Enable OpenShift BuildConfig and ImageStream | `false` |
| `openshift.build.sourceRepo` | Git repository URL for source builds | `https://github.com/tinmancoding/calculator-as-a-service.git` |
| `openshift.build.sourceRef` | Git branch/tag/commit for builds | `main` |

### Per-Service Configuration

Each service (gateway, parser, addition, subtraction, multiplication, division) supports:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `<service>.enabled` | Enable/disable the service | `true` |
| `<service>.replicaCount` | Number of pod replicas | `2` |
| `<service>.image.repository` | Image repository name | varies |
| `<service>.image.tag` | Image tag | `latest` |
| `<service>.service.type` | Kubernetes service type | `ClusterIP` |
| `<service>.service.port` | Service port | varies |
| `<service>.resources` | CPU/Memory resource limits | configured |

### Example Custom Values

```yaml
# custom-values.yaml

# Use a different image registry
global:
  imageRegistry: myregistry.io/calculator

# Scale down for development
gateway:
  replicaCount: 1
parser:
  replicaCount: 1
addition:
  replicaCount: 1
subtraction:
  replicaCount: 1
multiplication:
  replicaCount: 1
division:
  replicaCount: 1

# Expose gateway as NodePort
gateway:
  service:
    type: NodePort
```

## OpenShift Build Configuration

When `openshift.build.enabled=true`, the chart creates:

1. **ImageStreams**: Local image repositories for each service
2. **BuildConfigs**: Build configurations that:
   - Pull source code from the specified Git repository
   - Build Docker images using the Dockerfile in each service directory
   - Push images to the local ImageStream
   - Trigger automatic deployments when builds complete

### Triggering Builds

```bash
# Trigger all builds
oc start-build calculator-gateway
oc start-build calculator-parser
oc start-build calculator-addition
oc start-build calculator-subtraction
oc start-build calculator-multiplication
oc start-build calculator-division

# Monitor build progress
oc get builds -w
oc logs -f bc/calculator-gateway
```

## Usage

### Accessing the Gateway

After installation, you can access the gateway service:

```bash
# Port-forward (for ClusterIP services)
kubectl port-forward svc/gateway-service 8080:8080

# Test the calculator
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "10 + 5 * 2"}'
```

### Example Requests

```bash
# Simple addition
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "5 + 3"}'

# Complex expression
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "(10 + 5) * 2 - 8 / 4"}'
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade calculator ./helm/calculator --values new-values.yaml

# Upgrade to a new chart version
helm upgrade calculator ./helm/calculator
```

## Uninstallation

```bash
# Uninstall the release
helm uninstall calculator

# If using a specific namespace
helm uninstall calculator --namespace calculator

# Delete the namespace (optional)
kubectl delete namespace calculator
```

## Monitoring

### Check Deployment Status

```bash
# View all pods
kubectl get pods -l app=calculator

# View services
kubectl get services -l app=calculator

# View deployments
kubectl get deployments -l app=calculator

# Check logs for a specific service
kubectl logs -l service=gateway -f
```

### Health Checks

All services include liveness and readiness probes:

```bash
# Check pod health
kubectl describe pod <pod-name>

# Port-forward and check health endpoint
kubectl port-forward svc/gateway-service 8080:8080
curl http://localhost:8080/health
```

## Architecture

```
┌─────────────┐
│   Gateway   │ :8080
└──────┬──────┘
       │
       ├──────> Parser :8081
       │
       ├──────> Addition :8082 ──┐
       │                          │
       ├──────> Subtraction :8083 ├──> (Inter-service calls)
       │                          │
       ├──────> Multiplication :8084 ┘
       │
       └──────> Division :8086
```

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -l app=calculator

# View pod events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>
```

### OpenShift builds failing

```bash
# Check build status
oc get builds

# View build logs
oc logs build/<build-name>

# Check BuildConfig
oc describe bc/<buildconfig-name>
```

### Services can't communicate

```bash
# Check ConfigMap
kubectl get configmap calculator-config -o yaml

# Verify DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup gateway-service

# Check network policies (if any)
kubectl get networkpolicies
```

## Development

To modify and test the chart locally:

```bash
# Lint the chart
helm lint ./helm/calculator

# Dry-run to see generated manifests
helm install calculator ./helm/calculator --dry-run --debug

# Template without installing
helm template calculator ./helm/calculator

# Test with specific values
helm template calculator ./helm/calculator --values test-values.yaml
```

## Contributing

To contribute to this Helm chart:

1. Make your changes
2. Test with `helm lint` and `helm install --dry-run`
3. Update documentation
4. Submit a pull request

## License

This Helm chart is part of the Calculator as a Service project.
