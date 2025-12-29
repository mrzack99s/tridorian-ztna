# 🔐 SSL/TLS Certificate Management - Google-Managed

Complete guide for SSL/TLS certificate setup on GKE using **Google-Managed Certificates**.

---

## 📋 Table of Contents

1. [Google-Managed Certificates](#google-managed-certificates)
2. [Quick Setup](#quick-setup)
3. [Verification](#verification)
4. [Troubleshooting](#troubleshooting)

---

## 🎯 Why Google-Managed Certificates?

**Google-Managed Certificates** are the recommended solution for GKE:

**Pros:**
- ✅ Fully managed by Google
- ✅ Automatic renewal
- ✅ No additional components needed
- ✅ Free
- ✅ Integrated with GKE Gateway API
- ✅ No rate limits

**Cons:**
- ⚠️ Only works on GKE
- ⚠️ Takes 15-60 minutes to provision
- ⚠️ Requires DNS to be configured first

---

## 🔐 Google-Managed Certificates

### Prerequisites

1. **GKE Cluster** with Gateway API enabled
2. **DNS records** configured and propagated
3. **gcloud CLI** installed

### Setup

#### Method 1: Using Script (Recommended)

```bash
# Set environment variables
export GCP_PROJECT_ID=your-project-id
export DOMAIN=yourdomain.com
export CERT_METHOD=google-managed

# Run setup script
./k8s/setup-certificates.sh
```

#### Method 2: Manual Setup

```bash
# 1. Create certificate
gcloud certificate-manager certificates create tridorian-ztna-cert \
  --domains="api.yourdomain.com,auth.yourdomain.com,admin.yourdomain.com,backoffice.yourdomain.com,grpc.yourdomain.com" \
  --project=your-project-id

# 2. Create certificate map
gcloud certificate-manager maps create tridorian-ztna-certmap \
  --project=your-project-id

# 3. Add certificate to map
gcloud certificate-manager maps entries create tridorian-ztna-entry \
  --map="tridorian-ztna-certmap" \
  --certificates="tridorian-ztna-cert" \
  --hostname="*.yourdomain.com" \
  --project=your-project-id
```

### Gateway Configuration

The Gateway is already configured with the annotation:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tridorian-ztna-gateway
  annotations:
    networking.gke.io/certmap: "tridorian-ztna-certmap"
spec:
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
```

### Verification

```bash
# Check certificate status
gcloud certificate-manager certificates describe tridorian-ztna-cert

# Check certificate map
gcloud certificate-manager maps describe tridorian-ztna-certmap

# Check Gateway status
kubectl describe gateway tridorian-ztna-gateway -n tridorian-ztna
```

---

## 🚀 Quick Setup

```bash
# Set environment variables
export GCP_PROJECT_ID=trivpn-demo-prj
export DOMAIN=yourdomain.com

# Run setup script
./k8s/setup-certificates.sh
```

---

## ✅ Verification

### Test HTTPS Endpoints

```bash
# Test each domain
curl -I https://api.yourdomain.com/health
curl -I https://auth.yourdomain.com/health
curl -I https://admin.yourdomain.com/
curl -I https://backoffice.yourdomain.com/

# Check certificate details
openssl s_client -connect api.yourdomain.com:443 -servername api.yourdomain.com < /dev/null

# Verify certificate issuer
curl -vI https://api.yourdomain.com 2>&1 | grep -i issuer
```

### Check Certificate Status

```bash
# Check certificate status
gcloud certificate-manager certificates describe tridorian-ztna-cert

# Check provisioning status
gcloud certificate-manager certificates describe tridorian-ztna-cert \
  --format="value(managed.provisioningIssue,managed.status)"
```

---

## 🔧 Troubleshooting

### Certificate stuck in "PROVISIONING"

```bash
# Check DNS records
dig api.yourdomain.com

# Check certificate status
gcloud certificate-manager certificates describe tridorian-ztna-cert

# Common issues:
# 1. DNS not configured
# 2. DNS not propagated (wait 24-48 hours)
# 3. Domain verification failed
```

### Certificate shows "FAILED_NOT_VISIBLE"

```bash
# Ensure DNS records point to Gateway IP
GATEWAY_IP=$(kubectl get gateway tridorian-ztna-gateway \
  -n tridorian-ztna \
  -o jsonpath='{.status.addresses[0].value}')

echo "Gateway IP: $GATEWAY_IP"
dig api.yourdomain.com  # Should match Gateway IP
```

---

## 🔄 Certificate Renewal

- ✅ **Automatic** - Google handles renewal
- ✅ **No action required**
- ✅ **Renewed 30 days before expiry**

---

## 📚 Additional Resources

- [Google Certificate Manager](https://cloud.google.com/certificate-manager/docs)
- [Gateway API TLS](https://gateway-api.sigs.k8s.io/guides/tls/)
- [GKE Gateway Controller](https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api)

---

## ✨ Best Practices

1. ✅ **Configure DNS first** - Before requesting certificates
2. ✅ **Wait for propagation** - DNS changes can take 24-48 hours
3. ✅ **Monitor status** - Check certificate provisioning status
4. ✅ **Use wildcard certs** - For multiple subdomains (optional)
5. ✅ **Enable auto-renewal** - Automatic by default

---

**Status**: ✅ **Production Ready**

Google-Managed SSL/TLS certificates for all services! 🔐
