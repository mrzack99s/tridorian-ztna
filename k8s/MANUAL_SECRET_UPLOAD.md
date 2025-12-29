# Manual Secret Upload to GCP Secret Manager

Guide for manually uploading secrets to GCP Secret Manager.

---

## 📋 Required Secrets

You need to create **9 secrets** in GCP Secret Manager:

| Secret Name | Description | Source |
|-------------|-------------|--------|
| `ztna-private-key` | EdDSA private key for JWT signing | `keys/prod/private_key.pem` |
| `ztna-public-key` | EdDSA public key for JWT verification | `keys/prod/public_key.pem` |
| `ztna-db-username` | Database username | Your choice |
| `ztna-db-password` | Database password | Your choice |
| `ztna-db-name` | Database name | e.g., `tridorian_ztna_prod` |
| `ztna-db-host` | Database host | e.g., `postgres-service` |
| `ztna-cache-password` | Valkey/Redis password | Your choice |
| `ztna-oauth-google-client-id` | Google OAuth Client ID | From Google Console |
| `ztna-oauth-google-client-secret` | Google OAuth Client Secret | From Google Console |

---

## 🚀 Upload Methods

### Method 1: Using gcloud CLI (Recommended)

#### 1. EdDSA Keys

```bash
# Set project
export PROJECT_ID=trivpn-demo-prj

# Upload private key
gcloud secrets create ztna-private-key \
  --data-file=keys/prod/private_key.pem \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Upload public key
gcloud secrets create ztna-public-key \
  --data-file=keys/prod/public_key.pem \
  --replication-policy="automatic" \
  --project=$PROJECT_ID
```

#### 2. Database Credentials

```bash
# Username
echo -n "prod_user" | gcloud secrets create ztna-db-username \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Password (will prompt)
read -sp "Database Password: " DB_PASSWORD
echo -n "$DB_PASSWORD" | gcloud secrets create ztna-db-password \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Database name
echo -n "tridorian_ztna_prod" | gcloud secrets create ztna-db-name \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Database host
echo -n "postgres-service" | gcloud secrets create ztna-db-host \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID
```

#### 3. Cache Credentials

```bash
# Cache password (will prompt)
read -sp "Cache Password: " CACHE_PASSWORD
echo -n "$CACHE_PASSWORD" | gcloud secrets create ztna-cache-password \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID
```

#### 4. OAuth Credentials

```bash
# Google OAuth Client ID
echo -n "YOUR_CLIENT_ID.apps.googleusercontent.com" | \
  gcloud secrets create ztna-oauth-google-client-id \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Google OAuth Client Secret
echo -n "YOUR_CLIENT_SECRET" | \
  gcloud secrets create ztna-oauth-google-client-secret \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID
```

---

### Method 2: Using GCP Console (Web UI)

1. **Open GCP Console**
   - Navigate to: https://console.cloud.google.com/security/secret-manager
   - Select project: `trivpn-demo-prj`

2. **Create Secret**
   - Click "CREATE SECRET"
   - Enter secret name (e.g., `ztna-private-key`)
   - Paste or upload secret value
   - Replication: "Automatic"
   - Click "CREATE SECRET"

3. **Repeat for all 9 secrets**

---

## ✅ Verification

### List All Secrets

```bash
gcloud secrets list --project=trivpn-demo-prj --filter="name:ztna-*"
```

Expected output:
```
NAME                              CREATED              REPLICATION_POLICY  LOCATIONS
ztna-cache-password              2025-12-29T04:00:00  automatic           -
ztna-db-host                     2025-12-29T04:00:00  automatic           -
ztna-db-name                     2025-12-29T04:00:00  automatic           -
ztna-db-password                 2025-12-29T04:00:00  automatic           -
ztna-db-username                 2025-12-29T04:00:00  automatic           -
ztna-oauth-google-client-id      2025-12-29T04:00:00  automatic           -
ztna-oauth-google-client-secret  2025-12-29T04:00:00  automatic           -
ztna-private-key                 2025-12-29T04:00:00  automatic           -
ztna-public-key                  2025-12-29T04:00:00  automatic           -
```

### Verify Secret Value

```bash
# View secret metadata
gcloud secrets describe ztna-private-key --project=trivpn-demo-prj

# Get secret value (requires permissions)
gcloud secrets versions access latest \
  --secret=ztna-public-key \
  --project=trivpn-demo-prj
```

---

## 🔄 Update Existing Secret

```bash
# Add new version to existing secret
echo -n "new-value" | gcloud secrets versions add ztna-db-password \
  --data-file=- \
  --project=trivpn-demo-prj

# Or from file
gcloud secrets versions add ztna-private-key \
  --data-file=keys/prod/private_key.pem \
  --project=trivpn-demo-prj
```

---

## 🗑️ Delete Secret

```bash
# Delete secret (careful!)
gcloud secrets delete ztna-my-secret --project=trivpn-demo-prj
```

---

## 📊 Secret Checklist

Use this checklist to ensure all secrets are created:

- [ ] `ztna-private-key` - From `keys/prod/private_key.pem`
- [ ] `ztna-public-key` - From `keys/prod/public_key.pem`
- [ ] `ztna-db-username` - Database username
- [ ] `ztna-db-password` - Database password
- [ ] `ztna-db-name` - Database name
- [ ] `ztna-db-host` - Database host
- [ ] `ztna-cache-password` - Cache password
- [ ] `ztna-oauth-google-client-id` - OAuth client ID
- [ ] `ztna-oauth-google-client-secret` - OAuth client secret

---

## 🔐 Security Best Practices

1. ✅ **Never commit secrets to git**
2. ✅ **Use strong passwords** (min 16 characters)
3. ✅ **Enable audit logging**
4. ✅ **Restrict IAM access**
5. ✅ **Rotate secrets regularly** (every 90 days)
6. ✅ **Use different secrets per environment**

---

## 🚀 Next Steps

After uploading all secrets:

1. **Setup External Secrets Operator**
   ```bash
   ./k8s/setup-external-secrets.sh
   ```

2. **Apply SecretStore**
   ```bash
   kubectl apply -f k8s/base/secretstore.yaml
   ```

3. **Apply ExternalSecrets**
   ```bash
   kubectl apply -f k8s/base/externalsecrets.yaml
   ```

4. **Verify Sync**
   ```bash
   kubectl get externalsecret -n tridorian-ztna
   kubectl get secret -n tridorian-ztna
   ```

---

## 📚 Additional Resources

- [GCP Secret Manager Documentation](https://cloud.google.com/secret-manager/docs)
- [gcloud secrets commands](https://cloud.google.com/sdk/gcloud/reference/secrets)
- [Secret Manager IAM](https://cloud.google.com/secret-manager/docs/access-control)

---

**Status**: ✅ **Ready for Manual Upload**

Upload all 9 secrets to GCP Secret Manager before deploying! 🔐

gcloud beta container clusters update triztna-dev-cluster \
    --location=asia-southeast1 \
    --enable-secret-sync