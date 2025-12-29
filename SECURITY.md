# 🔐 Secure Key Management - No GitHub Leaks!

## ✅ Problem Solved

**Before**: Private keys were hardcoded in source code → GitHub would detect as leaked secrets ❌

**After**: Keys are loaded from environment variables → Safe from GitHub detection ✅

---

## 🚀 Quick Start

### 1. Initial Setup
```bash
# One command to set everything up
make setup
```

This will:
- ✅ Copy `.env.example` to `.env`
- ✅ Generate new EdDSA key pair
- ✅ Update `.env` with generated keys
- ✅ Set proper file permissions

### 2. Start Services
```bash
# Load environment variables and start
source .env && make docker-up
```

Or use docker-compose directly:
```bash
docker-compose --env-file .env -f docker-compose.dev.yaml up -d
```

---

## 🔑 How It Works

### Key Generation
```bash
# Generate new keys anytime
make generate-keys
```

This creates:
- `keys/private_key.pem` - Private key (gitignored)
- `keys/public_key.pem` - Public key (gitignored)
- Updates `.env` with both keys

### Environment Variables
Keys are stored in `.env` file:
```bash
ZTNA_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----"

ZTNA_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
...
-----END PUBLIC KEY-----"
```

### Code Usage
All services load keys from environment:
```go
privPEM := utils.GetEnv("ZTNA_PRIVATE_KEY", "")
if privPEM == "" {
    log.Fatal("❌ ZTNA_PRIVATE_KEY required. Run: make setup")
}
```

---

## 🔒 Security Features

### ✅ What's Protected

1. **No Hardcoded Keys**
   - ❌ No private keys in source code
   - ❌ No keys in git history
   - ✅ All keys in `.env` (gitignored)

2. **GitHub Secret Scanning**
   - ✅ Won't detect any leaked secrets
   - ✅ Safe to commit all code
   - ✅ Keys only in local `.env`

3. **File Permissions**
   - `private_key.pem`: 600 (owner read/write only)
   - `public_key.pem`: 644 (world readable)
   - `.env`: Should be 600 (set manually if needed)

### ⚠️ What's Gitignored

```gitignore
# Environment files
.env
.env.local
.env.*.local

# Key files
*.pem
keys/private_key.pem

# Exception: public key CAN be committed (but we don't)
!keys/public_key.pem
```

---

## 📋 Development Workflow

### First Time Setup
```bash
# 1. Clone repository
git clone <repo>
cd tridorian-ztna

# 2. Run setup
make setup

# 3. Review .env file
cat .env

# 4. Start services
source .env && make docker-up
```

### Daily Development
```bash
# Load environment variables
source .env

# Run services locally
make run-management  # Terminal 1
make run-controlplane # Terminal 2
make run-auth        # Terminal 3
```

### Key Rotation
```bash
# Generate new keys
make generate-keys

# Restart services to use new keys
make docker-down
source .env && make docker-up
```

---

## 🏭 Production Deployment

### Option 1: Environment Variables (Recommended)
```bash
# Set in your deployment platform
export ZTNA_PRIVATE_KEY="$(cat /secure/path/private_key.pem)"
export ZTNA_PUBLIC_KEY="$(cat /secure/path/public_key.pem)"
```

### Option 2: Secrets Management
```bash
# AWS Secrets Manager
export ZTNA_PRIVATE_KEY=$(aws secretsmanager get-secret-value \
  --secret-id ztna/private-key --query SecretString --output text)

# HashiCorp Vault
export ZTNA_PRIVATE_KEY=$(vault kv get -field=private_key secret/ztna)

# Kubernetes Secrets
kubectl create secret generic ztna-keys \
  --from-file=private-key=./keys/private_key.pem \
  --from-file=public-key=./keys/public_key.pem
```

### Option 3: Docker Secrets
```yaml
# docker-compose.prod.yaml
services:
  management-api:
    environment:
      - ZTNA_PRIVATE_KEY_FILE=/run/secrets/ztna_private_key
      - ZTNA_PUBLIC_KEY_FILE=/run/secrets/ztna_public_key
    secrets:
      - ztna_private_key
      - ztna_public_key

secrets:
  ztna_private_key:
    file: ./keys/private_key.pem
  ztna_public_key:
    file: ./keys/public_key.pem
```

---

## 🧪 Testing

### Verify Keys Are Loaded
```bash
# Start a service
source .env && make run-management

# Should see:
# 🔑 Public Key (PEM):
# -----BEGIN PUBLIC KEY-----
# ...
# -----END PUBLIC KEY-----
```

### Verify No Hardcoded Keys
```bash
# Search for private keys in code (should find none)
grep -r "BEGIN PRIVATE KEY" cmd/ internal/

# Should return nothing or only comments
```

### Verify .env Is Gitignored
```bash
# Check git status
git status

# .env should NOT appear in untracked files
```

---

## 📚 Files Overview

| File | Purpose | Git Status |
|------|---------|------------|
| `.env` | Contains actual keys | ❌ Gitignored |
| `.env.example` | Template without keys | ✅ Committed |
| `keys/private_key.pem` | Private key file | ❌ Gitignored |
| `keys/public_key.pem` | Public key file | ❌ Gitignored |
| `scripts/generate-keys.sh` | Key generation script | ✅ Committed |
| `cmd/*/main.go` | Load keys from env | ✅ Committed |

---

## ⚠️ Important Notes

### DO ✅
- Use `make setup` for initial setup
- Keep `.env` file secure (never commit)
- Rotate keys regularly (every 90 days)
- Use different keys per environment
- Use secrets management in production

### DON'T ❌
- Commit `.env` file to git
- Hardcode keys in source code
- Share private keys via email/chat
- Use same keys across environments
- Store keys in plain text in production

---

## 🔄 Migration from Old Setup

If you had hardcoded keys before:

```bash
# 1. Remove old .env if exists
rm -f .env

# 2. Run setup to generate new keys
make setup

# 3. Verify no hardcoded keys in code
grep -r "BEGIN PRIVATE KEY" cmd/

# 4. Commit the changes
git add .
git commit -m "feat: migrate to environment-based key management"
```

---

## 🐛 Troubleshooting

### Error: "ZTNA_PRIVATE_KEY environment variable is required"
```bash
# Solution: Run setup
make setup

# Or manually generate keys
make generate-keys

# Then load environment
source .env
```

### Error: "failed to parse private key"
```bash
# Solution: Regenerate keys
make generate-keys
source .env
```

### Keys not loading in Docker
```bash
# Solution: Pass .env file to docker-compose
docker-compose --env-file .env -f docker-compose.dev.yaml up -d

# Or export variables first
export $(cat .env | xargs)
docker-compose up -d
```

---

## ✅ Security Checklist

- [x] No private keys in source code
- [x] `.env` file is gitignored
- [x] Keys generated with proper permissions
- [x] Script to generate new keys
- [x] Clear documentation
- [x] Production-ready secrets management options
- [x] GitHub secret scanning won't detect leaks

---

**Status**: ✅ **SECURE & PRODUCTION-READY**

No more GitHub secret leak warnings! 🎉
