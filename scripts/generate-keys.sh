#!/bin/bash

# Generate EdDSA (Ed25519) Key Pair for Tridorian ZTNA
# This script generates new cryptographic keys for JWT signing

set -e

KEYS_DIR="./keys"
PRIVATE_KEY_FILE="$KEYS_DIR/private_key.pem"
PUBLIC_KEY_FILE="$KEYS_DIR/public_key.pem"
ENV_FILE=".env"

echo "🔐 Generating EdDSA (Ed25519) Key Pair..."

# Create keys directory if it doesn't exist
mkdir -p "$KEYS_DIR"

# Generate private key
echo "📝 Generating private key..."
openssl genpkey -algorithm ED25519 -out "$PRIVATE_KEY_FILE"

# Extract public key
echo "📝 Extracting public key..."
openssl pkey -in "$PRIVATE_KEY_FILE" -pubout -out "$PUBLIC_KEY_FILE"

# Set proper permissions
chmod 600 "$PRIVATE_KEY_FILE"
chmod 644 "$PUBLIC_KEY_FILE"

echo ""
echo "✅ Keys generated successfully!"
echo ""
echo "📁 Files created:"
echo "   - Private Key: $PRIVATE_KEY_FILE (600)"
echo "   - Public Key:  $PUBLIC_KEY_FILE (644)"
echo ""

# Read the keys
PRIVATE_KEY=$(cat "$PRIVATE_KEY_FILE")
PUBLIC_KEY=$(cat "$PUBLIC_KEY_FILE")

# Update or create .env file
echo "📝 Updating .env file..."

# Remove old key entries if they exist
if [ -f "$ENV_FILE" ]; then
    grep -v "^ZTNA_PRIVATE_KEY=" "$ENV_FILE" > "$ENV_FILE.tmp" || true
    grep -v "^ZTNA_PUBLIC_KEY=" "$ENV_FILE.tmp" > "$ENV_FILE" || true
    rm -f "$ENV_FILE.tmp"
fi

# Add new keys to .env
cat >> "$ENV_FILE" << EOF

# EdDSA (Ed25519) Keys for JWT Signing
# Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
ZTNA_PRIVATE_KEY="$PRIVATE_KEY"
ZTNA_PUBLIC_KEY="$PUBLIC_KEY"
EOF

echo "✅ .env file updated!"
echo ""
echo "⚠️  IMPORTANT SECURITY NOTES:"
echo "   1. Never commit .env file to git"
echo "   2. Keep private_key.pem secure"
echo "   3. Rotate keys regularly (every 90 days)"
echo "   4. Use different keys for dev/staging/prod"
echo ""
echo "🚀 Next steps:"
echo "   1. Copy .env.example to .env (if not exists)"
echo "   2. Run: source .env"
echo "   3. Start services: make docker-up"
echo ""
