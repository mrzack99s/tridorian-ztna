# GKE Secrets Store CSI Driver Integration

Complete guide for using GKE's native Secrets Store CSI Driver with GCP Secret Manager.

---

## 🎯 Overview

**GKE Secrets Store CSI Driver** is Google's native solution for mounting secrets from GCP Secret Manager directly into pods.

### Why CSI Driver over External Secrets Operator?

| Feature | CSI Driver | External Secrets Operator |
|---------|------------|---------------------------|
| **Native to GKE** | ✅ Built-in | ❌ Third-party |
| **Installation** | ✅ One command | ⚠️ Helm/kubectl |
| **Performance** | ✅ Direct mount | ⚠️ Sync delay |
| **Maintenance** | ✅ Google-managed | ⚠️ Self-managed |
| **Security** | ✅ No etcd storage | ⚠️ Stored in etcd |
| **Updates** | ✅ Real-time | ⚠️ Polling (1h default) |

---

## 🚀 Quick Start

### 1. Upload Secrets to GCP Secret Manager

See: [`MANUAL_SECRET_UPLOAD.md`](MANUAL_SECRET_UPLOAD.md)

```bash
# Upload all 9 required secrets
gcloud secrets create ztna-private-key \
  --data-file=keys/prod/private_key.pem \
  --project=trivpn-demo-prj

# ... (see MANUAL_SECRET_UPLOAD.md for complete list)
```

### 2. Enable CSI Driver on GKE Cluster

```bash
# Run setup script
./k8s/setup-gke-secrets-csi.sh

# Or manually:
gcloud container clusters update triztna-dev-cluster \
  --update-addons=GcpSecretManagerCsiDriver=ENABLED \
  --region=asia-southeast1 \
  --project=trivpn-demo-prj
```

### 3. Apply SecretProviderClass

```bash
kubectl apply -f k8s/base/secretproviderclass.yaml
```

### 4. Deploy Applications

```bash
# Secrets will be automatically mounted when pods start
kubectl apply -f k8s/base/
```

---

## 📊 Architecture

```
GCP Secret Manager
        ↓
   (Workload Identity)
        ↓
GKE Secrets Store CSI Driver
        ↓
SecretProviderClass
        ↓
   CSI Volume Mount
        ↓
Pod (files in /mnt/secrets-store)
        ↓
Kubernetes Secret (optional sync)
```

---

## 🔐 How It Works

### 1. SecretProviderClass

Defines which secrets to mount:

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: ztna-secrets
spec:
  provider: gcp
  parameters:
    secrets: |
      - resourceName: "projects/trivpn-demo-prj/secrets/ztna-private-key/versions/latest"
        path: "private-key"
  secretObjects:  # Optional: sync to K8s Secret
  - secretName: ztna-keys
    type: Opaque
    data:
    - objectName: "private-key"
      key: "private-key"
```

### 2. Pod Volume Mount

Mount secrets as files:

```yaml
spec:
  serviceAccountName: gke-secrets-sa
  containers:
  - name: app
    volumeMounts:
    - name: ztna-secrets
      mountPath: "/mnt/secrets-store"
      readOnly: true
  volumes:
  - name: ztna-secrets
    csi:
      driver: secrets-store-gke.csi.k8s.io
      readOnly: true
      volumeAttributes:
        secretProviderClass: "ztna-secrets"
```

### 3. Access Secrets

Secrets are available as files:

```bash
# Inside pod
cat /mnt/secrets-store/private-key
cat /mnt/db-secrets/password
```

Or as environment variables (via synced K8s Secret):

```yaml
env:
- name: ZTNA_PRIVATE_KEY
  valueFrom:
    secretKeyRef:
      name: ztna-keys
      key: private-key
```

---

## 📁 Secret Mapping

| GCP Secret | SecretProviderClass | Mount Path | K8s Secret |
|------------|---------------------|------------|------------|
| `ztna-private-key` | ztna-secrets | `/mnt/secrets-store/private-key` | ztna-keys |
| `ztna-public-key` | ztna-secrets | `/mnt/secrets-store/public-key` | ztna-keys |
| `ztna-db-username` | database-secrets | `/mnt/db-secrets/username` | database-credentials |
| `ztna-db-password` | database-secrets | `/mnt/db-secrets/password` | database-credentials |
| `ztna-db-name` | database-secrets | `/mnt/db-secrets/database` | database-credentials |
| `ztna-db-host` | database-secrets | `/mnt/db-secrets/host` | database-credentials |
| `ztna-cache-password` | cache-secrets | `/mnt/cache-secrets/password` | cache-credentials |
| `ztna-oauth-google-client-id` | oauth-secrets | `/mnt/oauth-secrets/google-client-id` | oauth-credentials |
| `ztna-oauth-google-client-secret` | oauth-secrets | `/mnt/oauth-secrets/google-client-secret` | oauth-credentials |

---

## 🔄 Secret Rotation

### Automatic Rotation

1. Update secret in GCP Secret Manager:
```bash
echo -n "new-password" | gcloud secrets versions add ztna-db-password \
  --data-file=- \
  --project=trivpn-demo-prj
```

2. **Restart pods** to pick up new version:
```bash
kubectl rollout restart deployment/management-api -n tridorian-ztna
```

**Note**: CSI Driver mounts secrets at pod start time. Pods must be restarted to get updated secrets.

### Rotation with Zero Downtime

```bash
# 1. Update secret in GCP
gcloud secrets versions add ztna-db-password --data-file=-

# 2. Rolling restart (zero downtime)
kubectl rollout restart deployment/management-api -n tridorian-ztna
kubectl rollout status deployment/management-api -n tridorian-ztna
```

---

## 🔍 Verification

### Check CSI Driver Status

```bash
# Check if CSI driver is enabled
gcloud container clusters describe triztna-dev-cluster \
  --region=asia-southeast1 \
  --project=trivpn-demo-prj \
  --format="value(addonsConfig.gcpSecretManagerCsiDriver.enabled)"
```

### Check SecretProviderClass

```bash
# List SecretProviderClasses
kubectl get secretproviderclass -n tridorian-ztna

# Describe
kubectl describe secretproviderclass ztna-secrets -n tridorian-ztna
```

### Check Mounted Secrets

```bash
# Exec into pod
kubectl exec -it deployment/management-api -n tridorian-ztna -- sh

# List mounted secrets
ls -la /mnt/secrets-store/
ls -la /mnt/db-secrets/
ls -la /mnt/cache-secrets/

# View secret content
cat /mnt/secrets-store/public-key
```

### Check Synced K8s Secrets

```bash
# List secrets
kubectl get secret -n tridorian-ztna

# View secret
kubectl get secret ztna-keys -n tridorian-ztna -o yaml
```

---

## 🛡️ Security Features

### Workload Identity

- ✅ **No service account keys** - Uses Workload Identity
- ✅ **IAM-based access** - Fine-grained permissions
- ✅ **Automatic rotation** - No manual key management

### Secret Protection

- ✅ **Not stored in etcd** - Secrets only in memory
- ✅ **Read-only mounts** - Cannot be modified
- ✅ **Pod-level isolation** - Each pod gets its own mount
- ✅ **Audit logging** - Track all secret access

---

## 🚨 Troubleshooting

### CSI Driver Not Enabled

```bash
# Check addon status
gcloud container clusters describe triztna-dev-cluster \
  --region=asia-southeast1 \
  --format="value(addonsConfig)"

# Enable if needed
gcloud container clusters update triztna-dev-cluster \
  --update-addons=GcpSecretManagerCsiDriver=ENABLED \
  --region=asia-southeast1
```

### Pod Fails to Mount Secrets

```bash
# Check pod events
kubectl describe pod <pod-name> -n tridorian-ztna

# Common issues:
# 1. Secret doesn't exist in GCP
gcloud secrets describe ztna-private-key

# 2. IAM permissions missing
gcloud projects get-iam-policy trivpn-demo-prj \
  --flatten="bindings[].members" \
  --filter="bindings.members:gke-secrets-sa"

# 3. Workload Identity not configured
kubectl describe sa gke-secrets-sa -n tridorian-ztna
```

### Secret Not Syncing to K8s Secret

```bash
# Check SecretProviderClass has secretObjects defined
kubectl get secretproviderclass ztna-secrets -n tridorian-ztna -o yaml

# Check pod is using the volume
kubectl get pod <pod-name> -n tridorian-ztna -o yaml | grep -A 10 volumes
```

---

## 📚 Configuration Files

| File | Purpose |
|------|---------|
| `setup-gke-secrets-csi.sh` | Enable CSI driver and setup Workload Identity |
| `base/secretproviderclass.yaml` | Define which secrets to mount |
| `base/management-api.yaml` | Example deployment with CSI volumes |

---

## 🆚 Comparison with External Secrets Operator

| Aspect | GKE CSI Driver | External Secrets Operator |
|--------|----------------|---------------------------|
| **Installation** | Built-in addon | Requires Helm/kubectl |
| **Maintenance** | Google-managed | Self-managed |
| **Secret Storage** | Memory only | Stored in etcd |
| **Update Mechanism** | Pod restart | Polling (1h default) |
| **Performance** | Direct mount | API sync |
| **Portability** | GKE only | Any Kubernetes |
| **Complexity** | Lower | Higher |

---

## ✅ Best Practices

1. ✅ **Use Workload Identity** - No service account keys
2. ✅ **Enable audit logging** - Track secret access
3. ✅ **Rotate secrets regularly** - Every 90 days
4. ✅ **Use latest version** - Always mount `/versions/latest`
5. ✅ **Restart pods after rotation** - Required for updates
6. ✅ **Monitor secret access** - Use Cloud Logging
7. ✅ **Test rotation process** - Before production

---

## 📖 Additional Resources

- [GKE Secrets Store CSI Driver](https://cloud.google.com/secret-manager/docs/secret-manager-managed-csi-component)
- [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/)
- [GCP Secret Manager](https://cloud.google.com/secret-manager/docs)

---

**Status**: ✅ **Production Ready**

Native GKE secret management with CSI Driver! 🔐
