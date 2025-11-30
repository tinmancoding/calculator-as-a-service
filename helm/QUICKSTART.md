# Quick Start Guide - Helm Deployment

This guide will help you quickly deploy Calculator as a Service using Helm.

## Prerequisites

Ensure you have:
- Helm 3.x installed (`helm version`)
- kubectl configured (`kubectl cluster-info`)
- Access to a Kubernetes or OpenShift cluster

## 1. Standard Kubernetes Deployment (5 minutes)

Deploy using pre-built Docker images from Docker Hub:

```bash
# Navigate to the project root
cd /path/to/calculator-as-a-service

# Install the chart
helm install calculator ./helm/calculator --namespace calculator --create-namespace

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=calculator -n calculator --timeout=300s

# Port-forward to access the gateway
kubectl port-forward -n calculator svc/gateway-service 8080:8080

# Test the calculator (in another terminal)
curl -X POST http://localhost:8080/calculate \
  -H "Content-Type: application/json" \
  -d '{"expression": "10 + 5 * 2"}'
```

## 2. OpenShift Deployment with Source Builds (10-15 minutes)

Deploy on OpenShift and build images from source:

```bash
# Login to OpenShift
oc login

# Create a new project
oc new-project calculator

# Install with build enabled
helm install calculator ./helm/calculator \
  --set openshift.build.enabled=true \
  --namespace calculator

# Monitor builds
oc get builds -w
# Wait for all builds to complete (this may take 5-10 minutes)

# Check deployment status
oc get pods

# Access via route (if you create one)
oc expose svc/gateway-service
oc get route gateway-service
```

## 3. Development Deployment (minimal resources)

For local development with minimal resources:

```bash
helm install calculator ./helm/calculator \
  --values ./helm/calculator/examples/dev-values.yaml \
  --namespace calculator-dev \
  --create-namespace
```

## 4. Production Deployment (high availability)

For production with high availability:

```bash
helm install calculator ./helm/calculator \
  --values ./helm/calculator/examples/production-values.yaml \
  --namespace calculator-prod \
  --create-namespace
```

## Common Commands

### Check Status
```bash
# List releases
helm list -n calculator

# Check release status
helm status calculator -n calculator

# View deployed resources
kubectl get all -n calculator -l app=calculator
```

### Upgrade
```bash
# Upgrade with new values
helm upgrade calculator ./helm/calculator \
  --namespace calculator \
  --values custom-values.yaml

# Upgrade with inline value changes
helm upgrade calculator ./helm/calculator \
  --namespace calculator \
  --set gateway.replicaCount=3
```

### Debugging
```bash
# View pod logs
kubectl logs -n calculator -l service=gateway -f

# Describe a pod
kubectl describe pod -n calculator <pod-name>

# Check configmap
kubectl get configmap -n calculator calculator-config -o yaml

# Test service connectivity
kubectl run -it --rm debug --image=busybox --restart=Never -n calculator -- sh
# Then inside the pod:
wget -O- http://gateway-service:8080/health
```

### Uninstall
```bash
# Remove the deployment
helm uninstall calculator -n calculator

# Delete the namespace
kubectl delete namespace calculator
```

## Troubleshooting

### Pods stuck in ImagePullBackOff
- Check if images exist in the registry
- For OpenShift builds, ensure builds completed successfully: `oc get builds`

### Services can't communicate
- Verify ConfigMap: `kubectl get cm calculator-config -o yaml -n calculator`
- Test DNS: `kubectl run -it --rm debug --image=busybox -n calculator -- nslookup gateway-service`

### High memory usage
- Adjust resource limits in values.yaml
- Reduce replica counts for development

## Next Steps

- Explore the [full README](./helm/calculator/README.md) for detailed configuration
- Check example values in `./helm/calculator/examples/`
- Set up monitoring and logging
- Configure ingress/routes for external access
- Implement CI/CD pipelines

## Example Calculations

Once deployed, test with these examples:

```bash
# Simple operations
curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" \
  -d '{"expression": "5 + 3"}'

curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" \
  -d '{"expression": "10 * 2"}'

curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" \
  -d '{"expression": "100 / 5"}'

# Complex expression
curl -X POST http://localhost:8080/calculate -H "Content-Type: application/json" \
  -d '{"expression": "(10 + 5) * 2 - 8 / 4"}'
```

## Support

For issues or questions:
- Check the [README](./helm/calculator/README.md)
- Review [SYSTEM-DESIGN.md](./SYSTEM-DESIGN.md) for architecture details
- Open an issue on GitHub
