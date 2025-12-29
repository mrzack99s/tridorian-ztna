# Docker Build and Push Guide

Complete guide for building and pushing Docker images to Google Artifact Registry.

---

## 🚀 Quick Start

### 1. Setup Artifact Registry

```bash
# Create repository (one-time setup)
make setup-artifact-registry

# Configure Docker authentication
make docker-auth
```

### 2. Build and Push All Images

```bash
# Build and push everything
make docker-build-push

# Or separately:
make docker-build  # Build all images
make docker-push   # Push all images
```

---

## 📦 Available Images

| Service | Image Name | Dockerfile |
|---------|------------|------------|
| Management API | `management-api:latest` | `Dockerfile.management-api` |
| Gateway Control Plane | `gateway-controlplane:latest` | `Dockerfile.gateway-controlplane` |
| Auth API | `auth-api:latest` | `Dockerfile.auth-api` |
| Gateway Agent | `gateway:latest` | `Dockerfile.gateway` |
| Tenant Admin | `tenant-admin:latest` | `Dockerfile.tenant-admin` |
| Backoffice | `backoffice:latest` | `Dockerfile.backoffice` |

---

## 🔧 Configuration

### Environment Variables

```bash
# Set in Makefile or override
export PROJECT_ID=trivpn-demo-prj
export REGION=asia-southeast1
export REPO_NAME=tridorian-ztna
export IMAGE_TAG=latest
```

### Full Image Paths

```
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/gateway-controlplane:latest
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/auth-api:latest
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/gateway:latest
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/tenant-admin:latest
asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/backoffice:latest
```

---

## 🛠️ Build Commands

### Build Individual Images

```bash
# Backend services
make docker-build-management
make docker-build-controlplane
make docker-build-auth
make docker-build-gateway

# Frontend services
make docker-build-tenant-admin
make docker-build-backoffice

# All images
make docker-build
```

### Build with Custom Tag

```bash
# Build with version tag
IMAGE_TAG=v1.0.0 make docker-build

# Build with commit hash
IMAGE_TAG=$(git rev-parse --short HEAD) make docker-build
```

---

## 📤 Push Commands

### Push Individual Images

```bash
# Backend services
make docker-push-management
make docker-push-controlplane
make docker-push-auth
make docker-push-gateway

# Frontend services
make docker-push-tenant-admin
make docker-push-backoffice

# All images
make docker-push
```

### Push with Custom Tag

```bash
IMAGE_TAG=v1.0.0 make docker-push
```

---

## 🔄 Complete Workflow

### Development Workflow

```bash
# 1. Make code changes
# 2. Build and push
make docker-build-push

# 3. Deploy to GKE
kubectl rollout restart deployment/management-api -n tridorian-ztna
```

### Production Release

```bash
# 1. Tag version
export IMAGE_TAG=v1.0.0

# 2. Build and push
make docker-build-push

# 3. Update Kubernetes manifests
# Edit k8s/overlays/prod/kustomization.yaml
# Change newTag to v1.0.0

# 4. Deploy
kubectl apply -k k8s/overlays/prod/
```

---

## 🏗️ Artifact Registry Setup

### Create Repository

```bash
# Using Makefile
make setup-artifact-registry

# Or manually
gcloud artifacts repositories create tridorian-ztna \
  --repository-format=docker \
  --location=asia-southeast1 \
  --description="Tridorian ZTNA Docker images" \
  --project=trivpn-demo-prj
```

### Configure Docker Authentication

```bash
# Using Makefile
make docker-auth

# Or manually
gcloud auth configure-docker asia-southeast1-docker.pkg.dev
```

### Verify Repository

```bash
# List repositories
gcloud artifacts repositories list --project=trivpn-demo-prj

# List images
gcloud artifacts docker images list \
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/tridorian-ztna
```

---

## 🔍 Verification

### Check Local Images

```bash
# List built images
docker images | grep tridorian-ztna

# Inspect image
docker inspect asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

### Check Remote Images

```bash
# List images in Artifact Registry
gcloud artifacts docker images list \
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/tridorian-ztna

# Get image details
gcloud artifacts docker images describe \
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

---

## 🐛 Troubleshooting

### Authentication Failed

```bash
# Re-authenticate
gcloud auth login
gcloud auth configure-docker asia-southeast1-docker.pkg.dev

# Check credentials
gcloud auth list
```

### Repository Not Found

```bash
# Create repository
make setup-artifact-registry

# Verify it exists
gcloud artifacts repositories describe tridorian-ztna \
  --location=asia-southeast1 \
  --project=trivpn-demo-prj
```

### Build Failed

```bash
# Check Docker is running
docker ps

# Check Dockerfile exists
ls -la Dockerfile.*

# Build with verbose output
docker build --progress=plain -t test -f Dockerfile.management-api .
```

### Push Failed

```bash
# Check authentication
gcloud auth print-access-token

# Check network
ping asia-southeast1-docker.pkg.dev

# Retry with full path
docker push asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

---

## 📊 Image Sizes

Estimated sizes after build:

| Image | Size |
|-------|------|
| management-api | ~25MB |
| gateway-controlplane | ~25MB |
| auth-api | ~30MB |
| gateway | ~25MB |
| tenant-admin | ~50MB |
| backoffice | ~50MB |

---

## 🔐 Security

### Image Scanning

```bash
# Scan image for vulnerabilities
gcloud artifacts docker images scan \
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

### Access Control

```bash
# Grant pull access to GKE
gcloud artifacts repositories add-iam-policy-binding tridorian-ztna \
  --location=asia-southeast1 \
  --member=serviceAccount:PROJECT_NUMBER-compute@developer.gserviceaccount.com \
  --role=roles/artifactregistry.reader \
  --project=trivpn-demo-prj
```

---

## 📚 Additional Resources

- [Artifact Registry Documentation](https://cloud.google.com/artifact-registry/docs)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)

---

**Status**: ✅ **Production Ready**

Complete Docker build and push workflow for Google Artifact Registry! 🐳
