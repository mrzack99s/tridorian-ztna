# Development Keys - Ed25519 (EdDSA)
# Generated: 2025-12-29T02:37:XX
# Environment: DEVELOPMENT
# Safe for local development only

## 🔧 Development Keys

**These are DEVELOPMENT keys** - Safe for local use, testing, and CI/CD.

- ✅ Safe to commit to git (if needed for team)
- ✅ Can be shared with team members
- ✅ Used for local development only
- ⚠️ **NEVER** use in production

---

## Private Key

**File**: `keys/dev/private_key.pem`

**Permissions**: 600 (owner read/write only)

---

## Public Key

**File**: `keys/dev/public_key.pem`

**Permissions**: 644 (world readable)

---

## Usage

### Local Development

```bash
# Load development keys
export ZTNA_PRIVATE_KEY=$(cat keys/dev/private_key.pem)
export ZTNA_PUBLIC_KEY=$(cat keys/dev/public_key.pem)

# Or use .env file
source .env
```

### Docker Compose

```bash
# Keys are automatically loaded from .env
docker-compose -f docker-compose.dev.yaml up
```

### Devcontainer

Keys are automatically loaded when you open the project in VS Code devcontainer.

---

## Regenerating Development Keys

```bash
# Run the setup script
make setup

# Or generate manually
./scripts/generate-keys.sh
```

---

## Security Notes

- ✅ These keys are for **development only**
- ✅ Safe to use on localhost
- ✅ Can be shared within development team
- ❌ **NEVER** use these keys in production
- ❌ **NEVER** use these keys in staging
- ❌ **NEVER** expose these keys to the internet

---

**Environment**: Development  
**Status**: ✅ Active  
**Safe to Share**: Yes (within team)
