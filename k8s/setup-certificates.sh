#!/bin/bash

# Setup Google-Managed SSL Certificates for GKE Gateway
# This script creates certificates using Google Certificate Manager

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_ID="${GCP_PROJECT_ID:-trivpn-demo-prj}"
DOMAIN="${DOMAIN:-yourdomain.com}"

echo -e "${GREEN}🔐 Setting up Google-Managed SSL Certificates${NC}"
echo "=============================================="
echo "Project: $PROJECT_ID"
echo "Domain: $DOMAIN"
echo ""

# Domain list
DOMAINS=(
    "mgmt.$DOMAIN"
    "console.$DOMAIN"
    "beadmin.$DOMAIN"
    "gwapi.$DOMAIN"
)

echo -e "${YELLOW}📜 Creating Google-Managed Certificate...${NC}"

# Create certificate
DOMAIN_LIST=$(IFS=,; echo "${DOMAINS[*]}")

echo "Creating certificate for domains: $DOMAIN_LIST"
gcloud certificate-manager certificates create tridorian-ztna-cert \
    --domains="$DOMAIN_LIST" \
    --project="$PROJECT_ID" || echo "Certificate may already exist"

# Create certificate map
echo ""
echo -e "${YELLOW}📋 Creating Certificate Map...${NC}"
gcloud certificate-manager maps create tridorian-ztna-certmap \
    --project="$PROJECT_ID" || echo "Certificate map may already exist"

# Create certificate map entry for each domain
echo ""
echo -e "${YELLOW}🔗 Adding Certificate Map Entries...${NC}"
for domain in "${DOMAINS[@]}"; do
    echo "Adding entry for $domain..."
    ENTRY_NAME="${domain//./-}-entry"
    gcloud certificate-manager maps entries create "$ENTRY_NAME" \
        --map="tridorian-ztna-certmap" \
        --certificates="tridorian-ztna-cert" \
        --hostname="$domain" \
        --project="$PROJECT_ID" || echo "Entry may already exist"
done

echo ""
echo -e "${GREEN}✅ Google-Managed Certificates configured!${NC}"
echo ""
echo "⚠️  Important:"
echo "1. DNS records must be configured first"
echo "2. Point all domains to Gateway IP"
echo "3. Certificate provisioning takes 15-60 minutes"
echo "4. Gateway annotation is already set in gateway.yaml"
echo ""

# Show certificate status
echo -e "${YELLOW}📊 Certificate Status:${NC}"
gcloud certificate-manager certificates describe tridorian-ztna-cert \
    --project="$PROJECT_ID" || true

echo ""
echo -e "${YELLOW}🌍 Next Steps:${NC}"
echo "1. Deploy Gateway: kubectl apply -f k8s/gateway-api/gateway.yaml"
echo "2. Get Gateway IP: kubectl get gateway tridorian-ztna-gateway -n tridorian-ztna"
echo "3. Configure DNS A records to point to Gateway IP"
echo "4. Wait for certificate provisioning (15-60 minutes)"
echo "5. Verify: gcloud certificate-manager certificates describe tridorian-ztna-cert"
echo ""
