# PowerShell Build and Push Guide

Guide for using PowerShell scripts to build and push Docker images on Windows.

---

## 🚀 Quick Start

### Prerequisites

1. **Docker Desktop** - Install and ensure it's running
2. **Google Cloud SDK** - Install gcloud CLI
3. **PowerShell** - Windows PowerShell 5.1+ or PowerShell Core 7+

### Basic Usage

```powershell
# Build and push all images
.\build-and-push.ps1

# Build only (no push)
.\build-and-push.ps1 -BuildOnly

# Push only (assumes images are built)
.\build-and-push.ps1 -PushOnly

# Build and push specific service
.\build-and-push.ps1 -Service management-api
```

---

## 📦 Available Services

- `management-api` - Management API
- `gateway-controlplane` - Gateway Control Plane
- `auth-api` - Authentication API
- `tenant-admin` - Tenant Admin Frontend
- `backoffice` - Backoffice Frontend
- `all` - All services (default)

---

## 🔧 Parameters

### Project Configuration

```powershell
# Custom project ID
.\build-and-push.ps1 -ProjectId "my-project-id"

# Custom region
.\build-and-push.ps1 -Region "us-central1"

# Custom repository name
.\build-and-push.ps1 -RepoName "my-repo"

# Custom image tag
.\build-and-push.ps1 -ImageTag "v1.0.0"
```

### Build Options

```powershell
# Build only (skip push)
.\build-and-push.ps1 -BuildOnly

# Push only (skip build)
.\build-and-push.ps1 -PushOnly

# Build specific service
.\build-and-push.ps1 -Service management-api

# Build with version tag
.\build-and-push.ps1 -ImageTag "v1.0.0"
```

### Combined Examples

```powershell
# Build all with custom tag
.\build-and-push.ps1 -ImageTag "v1.0.0"

# Build and push management-api only
.\build-and-push.ps1 -Service management-api

# Build all, don't push
.\build-and-push.ps1 -BuildOnly

# Push all (assumes built)
.\build-and-push.ps1 -PushOnly
```

---

## 🏗️ First-Time Setup

### 1. Install Prerequisites

```powershell
# Check Docker
docker --version

# Check gcloud
gcloud --version

# If not installed, download:
# Docker Desktop: https://www.docker.com/products/docker-desktop
# Google Cloud SDK: https://cloud.google.com/sdk/docs/install
```

### 2. Authenticate

```powershell
# Login to Google Cloud
gcloud auth login

# Set project
gcloud config set project trivpn-demo-prj

# Configure Docker for Artifact Registry
gcloud auth configure-docker asia-southeast1-docker.pkg.dev
```

### 3. Create Artifact Registry Repository

```powershell
# Using gcloud
gcloud artifacts repositories create triztna `
  --repository-format=docker `
  --location=asia-southeast1 `
  --description="Tridorian ZTNA Docker images" `
  --project=trivpn-demo-prj
```

---

## 🔄 Development Workflow

### Daily Development

```powershell
# 1. Make code changes
# 2. Build and test locally
.\build-and-push.ps1 -BuildOnly

# 3. Test image locally
docker run -it --rm asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest

# 4. Push to registry
.\build-and-push.ps1 -PushOnly

# 5. Deploy to GKE
kubectl rollout restart deployment/management-api -n tridorian-ztna
```

### Release Process

```powershell
# 1. Tag version
$version = "v1.0.0"

# 2. Build with version tag
.\build-and-push.ps1 -ImageTag $version

# 3. Update Kubernetes manifests
# Edit k8s/overlays/prod/kustomization.yaml

# 4. Deploy
kubectl apply -k k8s/overlays/prod/
```

---

## 🔍 Verification

### Check Local Images

```powershell
# List images
docker images | Select-String "triztna"

# Inspect image
docker inspect asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

### Check Remote Images

```powershell
# List images in Artifact Registry
gcloud artifacts docker images list `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna

# Get image details
gcloud artifacts docker images describe `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

---

## 🐛 Troubleshooting

### Docker Not Running

```powershell
# Check Docker status
docker ps

# If fails, start Docker Desktop
Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"

# Wait for Docker to start
Start-Sleep -Seconds 30
```

### Authentication Issues

```powershell
# Re-authenticate
gcloud auth login
gcloud auth configure-docker asia-southeast1-docker.pkg.dev

# Check credentials
gcloud auth list
```

### Build Failures

```powershell
# Build with verbose output
docker build --progress=plain -t test -f Dockerfile.management-api .

# Check Dockerfile exists
Get-ChildItem Dockerfile.*

# Clean Docker cache
docker system prune -a
```

### Push Failures

```powershell
# Check network
Test-NetConnection asia-southeast1-docker.pkg.dev -Port 443

# Retry push
docker push asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest

# Check quota
gcloud artifacts repositories describe triztna `
  --location=asia-southeast1 `
  --project=trivpn-demo-prj
```

---

## 📊 Script Features

### Progress Reporting

- ✅ Color-coded output (Success, Info, Error)
- ✅ Build summary
- ✅ Push summary
- ✅ Next steps guidance

### Error Handling

- ✅ Validates Docker is running
- ✅ Validates gcloud is installed
- ✅ Stops on build failures
- ✅ Detailed error messages

### Flexibility

- ✅ Build all or specific services
- ✅ Build-only or push-only modes
- ✅ Custom tags and configuration
- ✅ Parallel or sequential builds

---

## 📚 Additional Commands

### Clean Up

```powershell
# Remove local images
docker rmi asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest

# Remove all triztna images
docker images | Select-String "triztna" | ForEach-Object {
    $imageId = ($_ -split '\s+')[2]
    docker rmi $imageId
}

# Clean Docker system
docker system prune -a --volumes
```

### Image Management

```powershell
# Tag image
docker tag `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:v1.0.0

# Pull image
docker pull asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest

# Save image to file
docker save -o management-api.tar `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

---

## 🔐 Security

### Image Scanning

```powershell
# Scan for vulnerabilities
gcloud artifacts docker images scan `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest

# View scan results
gcloud artifacts docker images list-vulnerabilities `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna/management-api:latest
```

---

## 📖 Examples

### Example 1: Build All Services

```powershell
# Build all services with default settings
.\build-and-push.ps1
```

### Example 2: Build Specific Service

```powershell
# Build only management-api
.\build-and-push.ps1 -Service management-api -BuildOnly
```

### Example 3: Production Release

```powershell
# Build with version tag
.\build-and-push.ps1 -ImageTag "v1.0.0"

# Verify
gcloud artifacts docker images list `
  asia-southeast1-docker.pkg.dev/trivpn-demo-prj/triztna `
  --filter="tags:v1.0.0"
```

### Example 4: Quick Push

```powershell
# Push already-built images
.\build-and-push.ps1 -PushOnly
```

---

**Status**: ✅ **Ready for Windows Development**

Complete PowerShell workflow for Docker build and push! 🐳
