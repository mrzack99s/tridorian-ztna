# 🌐 Production Gateway Configuration - Port 443

## ✅ Updated Configuration

All services are now exposed on **port 443 (HTTPS)** through the Gateway API.

---

## 🔐 Gateway Architecture

### Single HTTPS Listener (Port 443)

```
Internet (Port 443)
        ↓
   Gateway API
        ↓
  ┌────────────────────────────────┐
  │  HTTPS Listener (Port 443)     │
  │  - TLS Termination             │
  │  - Certificate Management      │
  └────────────────────────────────┘
        ↓
  Hostname-Based Routing
        ↓
  ┌─────────────────┬─────────────────┬──────────────────┐
  │                 │                 │                  │
api.yourdomain.com  auth.yourdomain.com  admin.yourdomain.com
  │                 │                 │                  │
  ↓                 ↓                 ↓                  ↓
Management API    Auth API      Tenant Admin      Backoffice
(Port 8080)      (Port 8081)    (Port 80)        (Port 80)
```

---

## 📊 Service Endpoints (Production)

| External URL | Internal Service | Internal Port | Protocol |
|--------------|------------------|---------------|----------|
| `https://api.yourdomain.com` | management-api | 8080 | HTTP |
| `https://auth.yourdomain.com` | auth-api | 8081 | HTTP |
| `https://admin.yourdomain.com` | tenant-admin | 80 | HTTP |
| `https://backoffice.yourdomain.com` | backoffice | 80 | HTTP |
| `https://grpc.yourdomain.com` | gateway-controlplane | 5443 | gRPC |

**All external traffic uses port 443 (HTTPS)**

---

## 🔄 Traffic Flow

### Example: Management API Request

```
1. Client Request:
   https://api.yourdomain.com/api/v1/tenants
   ↓
2. Gateway (Port 443):
   - TLS Termination
   - Certificate Validation
   - Hostname: api.yourdomain.com
   ↓
3. HTTPRoute Matching:
   - Path: /api/v1/tenants
   - Backend: management-api:8080
   ↓
4. Internal Service:
   - management-api Pod
   - Port 8080 (HTTP)
   ↓
5. Response:
   - Pod → Gateway
   - Gateway → Client (HTTPS)
```

---

## 🛡️ Security Features

### TLS Termination at Gateway

- ✅ **Port 443** - All external traffic encrypted
- ✅ **TLS 1.2+** - Modern TLS versions only
- ✅ **Certificate Management** - Automatic via Google Certificate Manager
- ✅ **HTTP → HTTPS Redirect** - Port 80 redirects to 443

### Internal Communication

- ✅ **HTTP** - Internal cluster traffic (encrypted by GKE network)
- ✅ **ClusterIP** - Services not exposed externally
- ✅ **Network Policies** - Optional pod-to-pod restrictions

---

## 📝 Gateway Configuration

### Gateway Resource

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tridorian-ztna-gateway
spec:
  listeners:
  # Single HTTPS listener for all services
  - name: https
    protocol: HTTPS
    port: 443
    allowedRoutes:
      kinds:
      - HTTPRoute
      - GRPCRoute
  # HTTP listener for redirect
  - name: http
    protocol: HTTP
    port: 80
```

### HTTPRoute Example (Management API)

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: management-api-route
spec:
  parentRefs:
  - name: tridorian-ztna-gateway
    sectionName: https  # Port 443
  hostnames:
  - "api.yourdomain.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api/
    backendRefs:
    - name: management-api
      port: 8080  # Internal port
```

---

## 🌍 DNS Configuration

All domains point to the same Gateway IP on port 443:

```bash
# Get Gateway IP
GATEWAY_IP=$(kubectl get gateway tridorian-ztna-gateway \
  -n tridorian-ztna \
  -o jsonpath='{.status.addresses[0].value}')

# DNS A Records (all point to same IP)
api.yourdomain.com          A  <GATEWAY_IP>  # Port 443
auth.yourdomain.com         A  <GATEWAY_IP>  # Port 443
admin.yourdomain.com        A  <GATEWAY_IP>  # Port 443
backoffice.yourdomain.com   A  <GATEWAY_IP>  # Port 443
grpc.yourdomain.com         A  <GATEWAY_IP>  # Port 443
```

**Gateway handles routing based on hostname**

---

## 🔐 SSL Certificate Configuration

### Option 1: Google-Managed Certificates (Recommended)

```bash
# Create certificate for all domains
gcloud certificate-manager certificates create tridorian-ztna-cert \
  --domains="api.yourdomain.com,auth.yourdomain.com,admin.yourdomain.com,backoffice.yourdomain.com,grpc.yourdomain.com"

# Create certificate map
gcloud certificate-manager maps create tridorian-ztna-certmap

# Add certificate to map
gcloud certificate-manager maps entries create tridorian-ztna-entry \
  --map="tridorian-ztna-certmap" \
  --certificates="tridorian-ztna-cert" \
  --hostname="*.yourdomain.com"
```

Gateway annotation:
```yaml
metadata:
  annotations:
    networking.gke.io/certmap: "tridorian-ztna-certmap"
```

### Option 2: Wildcard Certificate

```bash
# Single wildcard certificate
gcloud certificate-manager certificates create tridorian-ztna-wildcard \
  --domains="*.yourdomain.com"
```

---

## 🧪 Testing

### Test HTTPS Endpoints

```bash
# Management API (Port 443)
curl https://api.yourdomain.com/health
curl https://api.yourdomain.com/version
curl https://api.yourdomain.com/api/v1/tenants

# Auth API (Port 443)
curl https://auth.yourdomain.com/health
curl https://auth.yourdomain.com/auth/mgmt/login

# Frontend (Port 443)
curl https://admin.yourdomain.com/
curl https://backoffice.yourdomain.com/
```

### Test HTTP → HTTPS Redirect

```bash
# Should redirect to HTTPS
curl -I http://api.yourdomain.com/health
# Expected: 301 Moved Permanently
# Location: https://api.yourdomain.com/health
```

### Test gRPC (Port 443)

```bash
# gRPC over HTTPS
grpcurl -d '{"node_id": "test"}' \
  grpc.yourdomain.com:443 \
  gateway.v1.GatewayService/RegisterGateway
```

---

## 📊 Port Summary

### External Ports (Internet-facing)

| Port | Protocol | Purpose |
|------|----------|---------|
| **443** | HTTPS | All services (Management, Auth, Frontend, gRPC) |
| **80** | HTTP | Redirect to HTTPS |

### Internal Ports (Cluster-only)

| Service | Port | Protocol |
|---------|------|----------|
| management-api | 8080 | HTTP |
| auth-api | 8081 | HTTP |
| gateway-controlplane | 5443 | gRPC |
| tenant-admin | 80 | HTTP |
| backoffice | 80 | HTTP |
| postgres | 5432 | TCP |
| valkey | 6379 | TCP |

---

## ✅ Benefits

### Single Port Configuration

- ✅ **Simplified Firewall** - Only port 443 needs to be open
- ✅ **Standard HTTPS** - No custom ports for users
- ✅ **Better Security** - All traffic encrypted
- ✅ **Easier Management** - Single certificate for all services

### Hostname-Based Routing

- ✅ **Clean URLs** - `api.yourdomain.com` vs `yourdomain.com:8080`
- ✅ **Flexible Routing** - Easy to add new services
- ✅ **Path-Based Rules** - Route by path within hostname
- ✅ **Traffic Splitting** - Canary deployments per service

---

## 🔄 Migration Notes

### From Development (Multiple Ports)

**Development**:
```
http://localhost:8080  → Management API
http://localhost:8081  → Auth API
http://localhost:3000  → Tenant Admin
```

**Production**:
```
https://api.yourdomain.com      → Management API (Port 443)
https://auth.yourdomain.com     → Auth API (Port 443)
https://admin.yourdomain.com    → Tenant Admin (Port 443)
```

### Client Configuration

Update client applications to use:
- ✅ **HTTPS** instead of HTTP
- ✅ **Port 443** (or omit port - default for HTTPS)
- ✅ **Hostname-based URLs** instead of port-based

---

## 📚 Additional Resources

- [Gateway API Docs](https://gateway-api.sigs.k8s.io/)
- [GKE Gateway Controller](https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api)
- [Google Certificate Manager](https://cloud.google.com/certificate-manager/docs)

---

**Status**: ✅ **Production Ready**

All services configured for port 443 with proper TLS termination! 🔐
