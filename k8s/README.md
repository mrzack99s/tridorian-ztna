# Tridorian ZTNA - GKE Deployment with Gateway API

Complete Kubernetes manifests for deploying Tridorian ZTNA to Google Kubernetes Engine using Gateway API.

---

## 📁 Directory Structure

```
k8s/
├── base/                          # Base Kubernetes resources
│   ├── namespace.yaml            # Namespace definition
│   ├── configmap.yaml            # Application configuration
│   ├── secrets.yaml              # Secrets template
│   ├── management-api.yaml       # Management API deployment
│   ├── gateway-controlplane.yaml # Gateway Control Plane deployment
│   ├── auth-api.yaml             # Auth API deployment
│   ├── frontend.yaml             # Frontend deployments
│   ├── database-cache.yaml       # PostgreSQL & Valkey StatefulSets
│   └── kustomization.yaml        # Kustomize base config
├── gateway-api/                   # Gateway API resources
│   ├── gateway.yaml              # Gateway definition
│   ├── httproutes.yaml           # HTTP routing rules
│   └── grpcroute.yaml            # gRPC routing rules
├── overlays/
│   ├── dev/                      # Development overlay
│   └── prod/                     # Production overlay
│       ├── kustomization.yaml    # Prod kustomization
│       └── replicas-patch.yaml   # Replica count patches
└── deploy-gke.sh                 # Deployment script
```

---

## 🚀 Quick Start

### Prerequisites

1. **GKE Cluster** with Gateway API enabled
2. **gcloud CLI** installed and configured
3. **kubectl** installed
4. **kustomize** installed

### Enable Gateway API on GKE

```bash
# Create cluster with Gateway API
gcloud container clusters create tridorian-ztna-cluster \
  --region=us-central1 \
  --gateway-api=standard \
  --num-nodes=3 \
  --machine-type=n2-standard-4

# Or update existing cluster
gcloud container clusters update tridorian-ztna-cluster \
  --region=us-central1 \
  --gateway-api=standard
```

### Upload Secrets to Google Secret Manager

```bash
# Upload production keys
gcloud secrets create ztna-private-key \
  --data-file=keys/prod/private_key.pem

gcloud secrets create ztna-public-key \
  --data-file=keys/prod/public_key.pem

# Upload database password
echo -n "YOUR_DB_PASSWORD" | gcloud secrets create ztna-db-password --data-file=-

# Upload cache password
echo -n "YOUR_CACHE_PASSWORD" | gcloud secrets create ztna-cache-password --data-file=-
```

### Deploy

```bash
# Set environment variables
export GCP_PROJECT_ID=your-project-id
export GKE_CLUSTER_NAME=tridorian-ztna-cluster
export GKE_REGION=us-central1
export DEPLOY_ENV=prod

# Run deployment script
./k8s/deploy-gke.sh
```

---

## 🌐 Gateway API Architecture

### Why Gateway API?

Gateway API is the **next-generation** Ingress for Kubernetes:

- ✅ **Role-oriented**: Separate concerns for infrastructure vs application teams
- ✅ **Expressive**: Rich routing capabilities (HTTP, gRPC, TCP, TLS)
- ✅ **Extensible**: Custom resources for advanced use cases
- ✅ **Portable**: Works across cloud providers
- ✅ **Type-safe**: Strong typing and validation

### Gateway API vs Ingress

| Feature | Ingress | Gateway API |
|---------|---------|-------------|
| **HTTP Routing** | ✅ Basic | ✅ Advanced |
| **gRPC Support** | ❌ Limited | ✅ Native |
| **TLS Termination** | ✅ Basic | ✅ Advanced |
| **Traffic Splitting** | ❌ No | ✅ Yes |
| **Header Manipulation** | ❌ Limited | ✅ Yes |
| **Multi-tenancy** | ❌ Difficult | ✅ Built-in |

---

## 📊 Deployed Resources

### Backend Services

| Service | Replicas (Prod) | Port | Protocol |
|---------|-----------------|------|----------|
| Management API | 5 | 8080 | HTTP |
| Gateway Control Plane | 5 | 5443 | gRPC |
| Auth API | 5 | 8081 | HTTP |
| Tenant Admin | 3 | 80 | HTTP |
| Backoffice | 3 | 80 | HTTP |

### Infrastructure

| Component | Type | Storage |
|-----------|------|---------|
| PostgreSQL | StatefulSet | 50Gi |
| Valkey | StatefulSet | 10Gi |

### Gateway API Resources

- **Gateway**: `tridorian-ztna-gateway`
- **HTTPRoutes**: 3 (management, auth, redirect)
- **GRPCRoute**: 1 (gateway-controlplane)

---

## 🔐 Security Features

### Pod Security

- ✅ **runAsNonRoot**: All containers run as non-root
- ✅ **readOnlyRootFilesystem**: Immutable root filesystem
- ✅ **Drop ALL capabilities**: Minimal privileges
- ✅ **Resource limits**: CPU and memory constraints

### Network Security

- ✅ **TLS termination** at Gateway
- ✅ **HTTP to HTTPS redirect**
- ✅ **ClusterIP services** (internal only)
- ✅ **Network policies** (optional)

### Secrets Management

- ✅ **Google Secret Manager** integration
- ✅ **Kubernetes Secrets** for runtime
- ✅ **No secrets in git**

---

## 🌍 DNS Configuration

After deployment, configure DNS records:

```bash
# Get Gateway IP
GATEWAY_IP=$(kubectl get gateway tridorian-ztna-gateway \
  -n tridorian-ztna \
  -o jsonpath='{.status.addresses[0].value}')

echo "Gateway IP: $GATEWAY_IP"
```

Create DNS A records:

```
api.yourdomain.com          A  <GATEWAY_IP>
auth.yourdomain.com         A  <GATEWAY_IP>
admin.yourdomain.com        A  <GATEWAY_IP>
backoffice.yourdomain.com   A  <GATEWAY_IP>
grpc.yourdomain.com         A  <GATEWAY_IP>
```

---

## 📜 SSL Certificates

### Option 1: Google-Managed Certificates

```yaml
# Add to gateway.yaml
metadata:
  annotations:
    networking.gke.io/certmap: "tridorian-ztna-certmap"
```

Create Certificate Map:

```bash
gcloud certificate-manager certificates create tridorian-ztna-cert \
  --domains="api.yourdomain.com,auth.yourdomain.com,admin.yourdomain.com,backoffice.yourdomain.com"

gcloud certificate-manager maps create tridorian-ztna-certmap

gcloud certificate-manager maps entries create tridorian-ztna-entry \
  --map="tridorian-ztna-certmap" \
  --certificates="tridorian-ztna-cert" \
  --hostname="*.yourdomain.com"
```

### Option 2: Let's Encrypt with cert-manager

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        gatewayHTTPRoute:
          parentRefs:
          - name: tridorian-ztna-gateway
            namespace: tridorian-ztna
EOF
```

---

## 📈 Monitoring

### Prometheus Metrics

All services expose metrics on port `9090`:

```bash
# Port-forward to access metrics
kubectl port-forward -n tridorian-ztna \
  deployment/management-api 9090:9090
```

### Health Checks

- **Liveness Probe**: Ensures container is running
- **Readiness Probe**: Ensures container is ready for traffic

### Logs

```bash
# View logs
kubectl logs -n tridorian-ztna -l app=management-api -f

# View all pods logs
kubectl logs -n tridorian-ztna --all-containers=true -f
```

---

## 🔄 Scaling

### Horizontal Pod Autoscaler

```bash
# Create HPA for Management API
kubectl autoscale deployment management-api \
  --namespace=tridorian-ztna \
  --cpu-percent=70 \
  --min=3 \
  --max=10
```

### Manual Scaling

```bash
# Scale deployment
kubectl scale deployment management-api \
  --namespace=tridorian-ztna \
  --replicas=10
```

---

## 🔧 Troubleshooting

### Check Gateway Status

```bash
kubectl describe gateway tridorian-ztna-gateway -n tridorian-ztna
```

### Check HTTPRoute Status

```bash
kubectl describe httproute management-api-route -n tridorian-ztna
```

### Check Pod Status

```bash
kubectl get pods -n tridorian-ztna
kubectl describe pod <pod-name> -n tridorian-ztna
```

### View Events

```bash
kubectl get events -n tridorian-ztna --sort-by='.lastTimestamp'
```

---

## 🧹 Cleanup

```bash
# Delete all resources
kubectl delete namespace tridorian-ztna

# Delete Gateway API resources
kubectl delete gateway tridorian-ztna-gateway -n tridorian-ztna
```

---

## 📚 Additional Resources

- [Gateway API Documentation](https://gateway-api.sigs.k8s.io/)
- [GKE Gateway Controller](https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api)
- [Kustomize Documentation](https://kustomize.io/)
- [Google Secret Manager](https://cloud.google.com/secret-manager/docs)

---

**Status**: ✅ Production Ready

All manifests are ready for GKE deployment with Gateway API! 🚀
