# Helm Chart - Package and Distribution Guide

## Packaging the Chart

### Create a Chart Package

```bash
# Navigate to the helm directory
cd /path/to/calculator-as-a-service

# Package the chart
helm package ./helm/calculator

# This creates: calculator-1.0.0.tgz
```

### Update Chart Version

When making changes, update the version in `Chart.yaml`:

```yaml
version: 1.0.1  # Increment this
appVersion: "1.0.1"  # Update if app version changed
```

## Distribution Methods

### 1. GitHub Releases

```bash
# Package the chart
helm package ./helm/calculator

# Create a GitHub release and attach calculator-1.0.0.tgz
# Users can then install directly from the release:
helm install calculator https://github.com/tinmancoding/calculator-as-a-service/releases/download/v1.0.0/calculator-1.0.0.tgz
```

### 2. Helm Repository (GitHub Pages)

```bash
# Create a gh-pages branch
git checkout --orphan gh-pages
git rm -rf .

# Package charts
helm package ./helm/calculator

# Create index
helm repo index . --url https://tinmancoding.github.io/calculator-as-a-service/

# Commit and push
git add .
git commit -m "Initial Helm repository"
git push origin gh-pages

# Users can then add your repo:
helm repo add calculator https://tinmancoding.github.io/calculator-as-a-service/
helm repo update
helm install my-calculator calculator/calculator
```

### 3. OCI Registry (Recommended for Modern Helm)

```bash
# Enable OCI support (Helm 3.8+)
export HELM_EXPERIMENTAL_OCI=1

# Login to registry
helm registry login ghcr.io -u USERNAME

# Package and push
helm package ./helm/calculator
helm push calculator-1.0.0.tgz oci://ghcr.io/tinmancoding

# Users install with:
helm install calculator oci://ghcr.io/tinmancoding/calculator --version 1.0.0
```

### 4. Local Directory (Development)

```bash
# Users can install directly from the directory
helm install calculator ./helm/calculator
```

## Installation Methods for Users

### From Local Directory
```bash
helm install calculator ./helm/calculator --namespace calculator --create-namespace
```

### From GitHub Repository (Raw)
```bash
# Clone the repo first
git clone https://github.com/tinmancoding/calculator-as-a-service.git
cd calculator-as-a-service
helm install calculator ./helm/calculator --namespace calculator --create-namespace
```

### From Packaged Release
```bash
# Download the package
curl -LO https://github.com/tinmancoding/calculator-as-a-service/releases/download/v1.0.0/calculator-1.0.0.tgz

# Install from package
helm install calculator calculator-1.0.0.tgz --namespace calculator --create-namespace
```

### From Helm Repository
```bash
# Add the repository
helm repo add calculator https://tinmancoding.github.io/calculator-as-a-service/
helm repo update

# Search for charts
helm search repo calculator

# Install
helm install my-calculator calculator/calculator --namespace calculator --create-namespace
```

### From OCI Registry
```bash
# Install directly from OCI registry
helm install calculator oci://ghcr.io/tinmancoding/calculator \
  --version 1.0.0 \
  --namespace calculator \
  --create-namespace
```

## Best Practices

### Versioning
- Use semantic versioning (MAJOR.MINOR.PATCH)
- Increment MAJOR for breaking changes
- Increment MINOR for new features
- Increment PATCH for bug fixes

### Testing Before Release
```bash
# Lint the chart
helm lint ./helm/calculator

# Dry-run installation
helm install test ./helm/calculator --dry-run --debug

# Template and review
helm template test ./helm/calculator > output.yaml

# Test all example values
helm template test ./helm/calculator --values ./helm/calculator/examples/dev-values.yaml
helm template test ./helm/calculator --values ./helm/calculator/examples/production-values.yaml
helm template test ./helm/calculator --values ./helm/calculator/examples/openshift-values.yaml
```

### Chart Signing (Optional)
```bash
# Generate a GPG key if you don't have one
gpg --gen-key

# Package and sign
helm package --sign ./helm/calculator --key 'Your Name' --keyring ~/.gnupg/secring.gpg

# This creates:
# - calculator-1.0.0.tgz
# - calculator-1.0.0.tgz.prov

# Users verify with:
helm install calculator calculator-1.0.0.tgz --verify --keyring ~/.gnupg/pubring.gpg
```

## Current State

The chart is currently:
- ✅ Created and validated
- ✅ Documented
- ✅ Tested with multiple configurations
- ✅ Ready for local installation

To publish:
1. Choose a distribution method above
2. Package the chart: `helm package ./helm/calculator`
3. Follow the steps for your chosen method
4. Update README with installation instructions

## Recommended Next Steps

1. **For Development/Workshop Use**: Keep as local directory installation
   ```bash
   helm install calculator ./helm/calculator
   ```

2. **For Production/Public Use**: Set up a Helm repository or use OCI registry
   - GitHub Pages is free and simple
   - OCI registries (ghcr.io, Docker Hub) are modern and efficient

3. **Document Installation**: Update the main README.md with chosen installation method
