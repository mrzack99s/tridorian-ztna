# Dockerfile Summary

This document provides an overview of all Dockerfiles in the project.

## Active Dockerfiles (Microservices)

### 1. Dockerfile.management-api
**Service**: Management API  
**Port**: 8080  
**Binary**: `management-api`  
**Purpose**: HTTP REST API for backoffice and tenant administration  
**Size**: ~36MB  

**Build**:
```bash
docker build -f Dockerfile.management-api -t tridorian-ztna/management-api .
```

**Run**:
```bash
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e CACHE_HOST=valkey \
  tridorian-ztna/management-api
```

---

### 2. Dockerfile.gateway-controlplane
**Service**: Gateway Control Plane  
**Port**: 5443  
**Binary**: `gateway-controlplane`  
**Purpose**: gRPC server for gateway orchestration and control  
**Size**: ~33MB  

**Build**:
```bash
docker build -f Dockerfile.gateway-controlplane -t tridorian-ztna/gateway-controlplane .
```

**Run**:
```bash
docker run -p 5443:5443 \
  -e DB_HOST=postgres \
  -e CACHE_HOST=valkey \
  tridorian-ztna/gateway-controlplane
```

---

### 3. Dockerfile.auth-api
**Service**: Authentication API  
**Port**: 8081  
**Binary**: `auth-api`  
**Purpose**: OAuth2 and user authentication service  
**Size**: ~35MB  

**Build**:
```bash
docker build -f Dockerfile.auth-api -t tridorian-ztna/auth-api .
```

**Run**:
```bash
docker run -p 8081:8081 \
  -e DB_HOST=postgres \
  -e CACHE_HOST=valkey \
  -v ./ip2country-v4.tsv:/root/ip2country-v4.tsv:ro \
  tridorian-ztna/auth-api
```

---

### 4. Dockerfile.gateway
**Service**: Gateway Agent  
**Port**: 6500/udp  
**Binary**: `gateway`  
**Purpose**: VPN gateway agent running on edge nodes  
**Size**: ~33MB  

**Build**:
```bash
docker build -f Dockerfile.gateway -t tridorian-ztna/gateway .
```

**Run**:
```bash
docker run -p 6500:6500/udp \
  -e NODE_ID=<uuid> \
  -e CONTROL_PLANE_ADDR=gateway-controlplane:5443 \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  tridorian-ztna/gateway
```

---

### 5. Dockerfile.tenant-admin
**Service**: Tenant Admin Frontend  
**Port**: 80 (mapped to 3000)  
**Binary**: N/A (Static files)  
**Purpose**: Web UI for tenant administrators  
**Size**: ~25MB  
**Tech**: React + Vite + Nginx

**Build**:
```bash
docker build -f Dockerfile.tenant-admin -t tridorian-ztna/tenant-admin .
```

**Run**:
```bash
docker run -p 3000:80 \
  -e VITE_API_URL=http://localhost:8080 \
  -e VITE_AUTH_URL=http://localhost:8081 \
  tridorian-ztna/tenant-admin
```

---

### 6. Dockerfile.backoffice
**Service**: Backoffice Frontend  
**Port**: 80 (mapped to 3001)  
**Binary**: N/A (Static files)  
**Purpose**: Web UI for system administrators  
**Size**: ~25MB  
**Tech**: React + Vite + Nginx

**Build**:
```bash
docker build -f Dockerfile.backoffice -t tridorian-ztna/backoffice .
```

**Run**:
```bash
docker run -p 3001:80 \
  -e VITE_API_URL=http://localhost:8080 \
  -e VITE_AUTH_URL=http://localhost:8081 \
  tridorian-ztna/backoffice
```

---

## Deprecated Dockerfiles

### 5. Dockerfile.tridorian-ztna (Legacy)
**Service**: Monolith (Deprecated)  
**Ports**: 8080, 5443  
**Binary**: `tridorian-ztna`  
**Purpose**: Legacy monolithic application (HTTP + gRPC in one)  
**Status**: ⚠️ **DEPRECATED** - Use microservices instead  

This Dockerfile is kept for backward compatibility but should not be used for new deployments.

---

## Build All Images

```bash
# Build all backend services
docker build -f Dockerfile.management-api -t tridorian-ztna/management-api .
docker build -f Dockerfile.gateway-controlplane -t tridorian-ztna/gateway-controlplane .
docker build -f Dockerfile.auth-api -t tridorian-ztna/auth-api .
docker build -f Dockerfile.gateway -t tridorian-ztna/gateway .

# Build all frontend services
docker build -f Dockerfile.tenant-admin -t tridorian-ztna/tenant-admin .
docker build -f Dockerfile.backoffice -t tridorian-ztna/backoffice .
```

Or use docker-compose:
```bash
make docker-build
```

---

## Image Sizes Comparison

| Service | Dockerfile | Binary/Build Size | Image Size |
|---------|-----------|-------------------|------------|
| Management API | `Dockerfile.management-api` | 36MB | ~50MB |
| Gateway Control Plane | `Dockerfile.gateway-controlplane` | 33MB | ~47MB |
| Authentication API | `Dockerfile.auth-api` | 35MB | ~50MB |
| Gateway Agent | `Dockerfile.gateway` | 19MB | ~35MB |
| Tenant Admin | `Dockerfile.tenant-admin` | N/A | ~25MB |
| Backoffice | `Dockerfile.backoffice` | N/A | ~25MB |
| Legacy Monolith | `Dockerfile.tridorian-ztna` | 43MB | ~60MB |

**Total (All Services)**: ~232MB  
**Backend Only**: ~182MB  
**Frontend Only**: ~50MB

---

## Multi-Stage Build Strategy

All Dockerfiles use a **multi-stage build** approach:

1. **Builder Stage** (`golang:1.23-alpine`):
   - Install build dependencies
   - Download Go modules
   - Compile the binary with `CGO_ENABLED=0`

2. **Runtime Stage** (`alpine:latest`):
   - Minimal base image
   - Only essential runtime dependencies (ca-certificates, tzdata)
   - Copy compiled binary from builder
   - No source code or build tools

This approach ensures:
- ✅ Small image sizes
- ✅ Fast deployment
- ✅ Reduced attack surface
- ✅ No unnecessary dependencies

---

## Environment Variables

All services support these common environment variables:

```env
# Application
APP_ENV=development|production

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=devuser
DB_PASSWORD=P@ssw0rd
DB_NAME=trivpn-trimanaged

# Cache
CACHE_HOST=localhost
CACHE_PORT=6379
CACHE_PASSWORD=P@ssw0rd
```

Service-specific variables:
- **Management API**: `MGMT_PORT=8080`
- **Gateway Control Plane**: `GRPC_PORT=5443`
- **Authentication API**: `AUTH_PORT=8081`
- **Gateway Agent**: `NODE_ID`, `CONTROL_PLANE_ADDR`, `VPN_PORT=6500`

---

## Production Considerations

### Security
1. **Don't use hardcoded keys** - Use secrets management
2. **Enable TLS** - All services should use TLS in production
3. **Use non-root user** - Add `USER` directive in Dockerfile
4. **Scan images** - Use tools like Trivy or Snyk

### Optimization
1. **Use specific tags** - Don't use `latest` in production
2. **Multi-arch builds** - Support ARM64 for cost savings
3. **Layer caching** - Order Dockerfile commands for optimal caching
4. **Health checks** - Add HEALTHCHECK directive

### Example Production Dockerfile Enhancement:
```dockerfile
# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as non-root
RUN adduser -D -u 1000 appuser
USER appuser
```

---

## CI/CD Integration

### GitHub Actions Example:
```yaml
- name: Build Docker Images
  run: |
    docker build -f Dockerfile.management-api -t ${{ env.REGISTRY }}/management-api:${{ github.sha }} .
    docker build -f Dockerfile.gateway-controlplane -t ${{ env.REGISTRY }}/gateway-controlplane:${{ github.sha }} .
    docker build -f Dockerfile.auth-api -t ${{ env.REGISTRY }}/auth-api:${{ github.sha }} .
    docker build -f Dockerfile.gateway -t ${{ env.REGISTRY }}/gateway:${{ github.sha }} .
```

---

## Troubleshooting

### Build fails with "module not found"
```bash
# Clear Go module cache
go clean -modcache
docker build --no-cache -f Dockerfile.xxx .
```

### Image size too large
```bash
# Check layers
docker history tridorian-ztna/management-api

# Use dive to analyze
dive tridorian-ztna/management-api
```

### Container won't start
```bash
# Check logs
docker logs <container-id>

# Run interactively
docker run -it --entrypoint /bin/sh tridorian-ztna/management-api
```

---

## References

- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Best Practices for Writing Dockerfiles](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Alpine Linux](https://alpinelinux.org/)
