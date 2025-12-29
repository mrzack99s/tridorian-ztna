# EdDSA Key Pair (Ed25519)

## 🔐 Generated Keys

**Generated on**: 2025-12-29T02:31:17Z  
**Algorithm**: Ed25519 (EdDSA)  
**Key Size**: 256 bits

---

## Private Key

```
-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBcWdXbUcLGY8u+oQ3iWs07HgSDe2J/yvY7s6Fmq/C1x
-----END PRIVATE KEY-----
```

---

## Public Key

```
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAV+Igx69pUElnsK+STUumH8zFrkykXKKUSPoVc3Wbdec=
-----END PUBLIC KEY-----
```

---

## ⚠️ Security Warning

**IMPORTANT**: These keys are embedded in the source code for **DEVELOPMENT ONLY**.

### For Production:

1. **DO NOT** commit private keys to version control
2. **DO NOT** hardcode keys in source code
3. **DO** use secure secrets management:
   - HashiCorp Vault
   - AWS Secrets Manager
   - Azure Key Vault
   - Google Cloud Secret Manager
   - Kubernetes Secrets

### Environment Variables

Set these environment variables in production:

```bash
# Management API
export MGMT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIBcWdXbUcLGY8u+oQ3iWs07HgSDe2J/yvY7s6Fmq/C1x
-----END PRIVATE KEY-----"

# Or load from file
export MGMT_PRIVATE_KEY=$(cat /secure/path/private_key.pem)
```

---

## 📍 Where Keys Are Used

### Private Key (JWT Signing)
- `cmd/management-api/main.go` - Signs JWT tokens for management API
- `cmd/auth-api/main.go` - Signs JWT tokens for authentication
- `cmd/triztna/main.go` - Legacy monolith

### Public Key (JWT Verification)
- `cmd/management-api/main.go` - Verifies JWT tokens
- `cmd/gateway-controlpane/main.go` - Verifies gateway authentication
- `cmd/auth-api/main.go` - Verifies JWT tokens
- `cmd/triztna/main.go` - Legacy monolith

---

## 🔄 Key Rotation

To rotate keys:

1. Generate new key pair:
   ```bash
   openssl genpkey -algorithm ED25519 -out new_private_key.pem
   openssl pkey -in new_private_key.pem -pubout -out new_public_key.pem
   ```

2. Update all services with new keys

3. Deploy services with zero-downtime strategy:
   - Deploy with both old and new public keys (verify both)
   - Wait for all tokens to expire or force re-authentication
   - Remove old public key

---

## 🧪 Testing Keys

To verify the key pair:

```bash
# Extract public key from private key
openssl pkey -in private_key.pem -pubout

# Should match the public key above
```

---

## 📚 Additional Information

### Ed25519 Benefits:
- ✅ Fast signature generation and verification
- ✅ Small key size (32 bytes)
- ✅ High security (128-bit security level)
- ✅ Deterministic signatures
- ✅ No random number generator needed for signing

### Use Cases:
- JWT token signing and verification
- API authentication
- Service-to-service authentication
- Gateway authentication

---

## 🔒 Best Practices

1. **Never share private keys**
2. **Rotate keys regularly** (every 90 days recommended)
3. **Use different keys for different environments** (dev, staging, prod)
4. **Monitor key usage** and audit access
5. **Have a key revocation plan**
6. **Backup keys securely** (encrypted backups)

---

## 📞 Emergency Key Revocation

If private key is compromised:

1. **Immediately** generate new key pair
2. **Deploy** new keys to all services
3. **Invalidate** all existing JWT tokens
4. **Force** all users to re-authenticate
5. **Audit** all API access during compromise period
6. **Document** the incident

---

**Last Updated**: 2025-12-29T02:31:17Z  
**Status**: ✅ Active (Development Only)
