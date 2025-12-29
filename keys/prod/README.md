# Production Keys - Ed25519 (EdDSA)
# Generated: 2025-12-29T02:43:10Z
# Environment: PRODUCTION
# Rotation Due: 2025-03-29

## ⚠️ CRITICAL SECURITY WARNING

**These are PRODUCTION keys!**

- **NEVER** commit this file to git
- **NEVER** share these keys via email/chat
- **ALWAYS** use secrets management in production
- **ROTATE** keys every 90 days

---

## Private Key

**File**: `keys/prod/private_key.pem`

```
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIGw0tWR75DGdXSzGXI/DJI0g5Zdo6EZSafOtTdby131+
-----END PRIVATE KEY-----
```

**Permissions**: 600 (owner read/write only)

---

## Public Key

**File**: `keys/prod/public_key.pem`

```
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEA9FGwZVmGTqoKW4w3AkqSDYw5zbsKz2F+jl18k3Rg+Q8=
-----END PUBLIC KEY-----
```

**Permissions**: 644 (world readable)

---

## Deployment Instructions

### Option 1: Environment Variables (Recommended)

```bash
# Load from file
export ZTNA_PRIVATE_KEY=$(cat keys/prod/private_key.pem)
export ZTNA_PUBLIC_KEY=$(cat keys/prod/public_key.pem)

# Or load entire .env.production
export $(cat .env.production | grep -v '^#' | xargs)
```

### Option 2: AWS Secrets Manager

```bash
# Store private key
aws secretsmanager create-secret \
  --name ztna/prod/private-key \
  --secret-string file://keys/prod/private_key.pem

# Store public key
aws secretsmanager create-secret \
  --name ztna/prod/public-key \
  --secret-string file://keys/prod/public_key.pem

# Retrieve in application
export ZTNA_PRIVATE_KEY=$(aws secretsmanager get-secret-value \
  --secret-id ztna/prod/private-key \
  --query SecretString --output text)
```

### Option 3: HashiCorp Vault

```bash
# Store keys
vault kv put secret/ztna/prod \
  private_key=@keys/prod/private_key.pem \
  public_key=@keys/prod/public_key.pem

# Retrieve in application
export ZTNA_PRIVATE_KEY=$(vault kv get -field=private_key secret/ztna/prod)
export ZTNA_PUBLIC_KEY=$(vault kv get -field=public_key secret/ztna/prod)
```

### Option 4: Kubernetes Secrets

```bash
# Create secret
kubectl create secret generic ztna-prod-keys \
  --from-file=private-key=keys/prod/private_key.pem \
  --from-file=public-key=keys/prod/public_key.pem \
  --namespace=production

# Use in deployment
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: management-api
    env:
    - name: ZTNA_PRIVATE_KEY
      valueFrom:
        secretKeyRef:
          name: ztna-prod-keys
          key: private-key
```

---

## Key Rotation Schedule

| Date | Action | Status |
|------|--------|--------|
| 2025-12-29 | Initial generation | ✅ Complete |
| 2026-03-29 | First rotation | ⏳ Scheduled |
| 2026-06-27 | Second rotation | ⏳ Scheduled |
| 2026-09-25 | Third rotation | ⏳ Scheduled |

**Rotation Frequency**: Every 90 days

---

## Security Checklist

- [x] Keys generated with Ed25519 algorithm
- [x] Private key has 600 permissions
- [x] Public key has 644 permissions
- [x] Keys stored in separate prod directory
- [x] .env.production created with keys
- [x] Keys NOT committed to git
- [ ] Keys uploaded to secrets management
- [ ] Production deployment configured
- [ ] Monitoring and alerts set up
- [ ] Key rotation calendar created

---

## Emergency Procedures

### If Private Key is Compromised

1. **Immediately** generate new keys:
   ```bash
   ./scripts/generate-keys.sh --env=production
   ```

2. **Update** secrets management:
   ```bash
   # AWS
   aws secretsmanager update-secret \
     --secret-id ztna/prod/private-key \
     --secret-string file://keys/prod/private_key.pem
   ```

3. **Deploy** new keys to all services

4. **Invalidate** all existing JWT tokens

5. **Force** all users to re-authenticate

6. **Audit** all access logs

7. **Document** the incident

---

## Contact Information

**Security Team**: security@yourdomain.com  
**On-Call**: +1-XXX-XXX-XXXX  
**Incident Response**: https://yourdomain.com/security/incident

---

**Last Updated**: 2025-12-29T02:43:10Z  
**Next Review**: 2026-03-29  
**Status**: ✅ Active (Production)
