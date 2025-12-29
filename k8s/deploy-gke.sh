#!/bin/bash

# Tridorian ZTNA - GKE Deployment Script
# This script deploys the application to Google Kubernetes Engine

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-trivpn-demo-prj}"
CLUSTER_NAME="${GKE_CLUSTER_NAME:-triztna-dev-cluster}"
REGION="${GKE_REGION:-asia-southeast1}"
ENVIRONMENT="${DEPLOY_ENV:-prod}"

echo -e "${GREEN}🚀 Tridorian ZTNA - GKE Deployment${NC}"
echo "=================================="
echo "Project ID: $PROJECT_ID"
echo "Cluster: $CLUSTER_NAME"
echo "Region: $REGION"
echo "Environment: $ENVIRONMENT"
echo ""

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"

if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI not found. Please install it first.${NC}"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found. Please install it first.${NC}"
    exit 1
fi

if ! command -v kustomize &> /dev/null; then
    echo -e "${RED}❌ kustomize not found. Please install it first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites met${NC}"
echo ""

# Set GCP project
echo -e "${YELLOW}🔧 Setting GCP project...${NC}"
gcloud config set project $PROJECT_ID

# Get cluster credentials
echo -e "${YELLOW}🔑 Getting cluster credentials...${NC}"
gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION

# Create namespace if it doesn't exist
echo -e "${YELLOW}📦 Creating namespace...${NC}"
kubectl apply -f k8s/base/namespace.yaml

# Create secrets from production keys
echo -e "${YELLOW}🔐 Creating secrets...${NC}"
echo "⚠️  Make sure you have uploaded keys to Google Secret Manager first!"
echo ""

# Option 1: Create from Google Secret Manager
if command -v gcloud &> /dev/null; then
    echo "Creating secrets from Google Secret Manager..."
    
    # Get private key from Secret Manager
    PRIVATE_KEY=$(gcloud secrets versions access latest --secret="ztna-private-key")
    PUBLIC_KEY=$(gcloud secrets versions access latest --secret="ztna-public-key")
    DB_PASSWORD=$(gcloud secrets versions access latest --secret="ztna-db-password")
    CACHE_PASSWORD=$(gcloud secrets versions access latest --secret="ztna-cache-password")
    
    # Create Kubernetes secrets
    kubectl create secret generic ztna-keys \
        --from-literal=private-key="$PRIVATE_KEY" \
        --from-literal=public-key="$PUBLIC_KEY" \
        --namespace=tridorian-ztna \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic database-credentials \
        --from-literal=username="prod_user" \
        --from-literal=password="$DB_PASSWORD" \
        --from-literal=database="tridorian_ztna_prod" \
        --namespace=tridorian-ztna \
        --dry-run=client -o yaml | kubectl apply -f -
    
    kubectl create secret generic cache-credentials \
        --from-literal=password="$CACHE_PASSWORD" \
        --namespace=tridorian-ztna \
        --dry-run=client -o yaml | kubectl apply -f -
fi

# Apply ConfigMap
echo -e "${YELLOW}⚙️  Applying ConfigMap...${NC}"
kubectl apply -f k8s/base/configmap.yaml

# Setup SSL/TLS Certificates
echo -e "${YELLOW}🔐 Setting up SSL/TLS Certificates...${NC}"
echo "Choose certificate method:"
echo "1) Google-Managed Certificates (recommended for GKE)"
echo "2) Let's Encrypt with cert-manager"
echo "3) Skip (configure manually later)"
read -p "Enter choice [1-3]: " cert_choice

case $cert_choice in
    1)
        echo "Setting up Google-Managed Certificates..."
        CERT_METHOD=google-managed ./k8s/setup-certificates.sh
        ;;
    2)
        echo "Setting up Let's Encrypt..."
        CERT_METHOD=letsencrypt ./k8s/setup-certificates.sh
        ;;
    3)
        echo "Skipping certificate setup"
        ;;
    *)
        echo "Invalid choice, skipping certificate setup"
        ;;
esac
echo ""

# Deploy using Kustomize
echo -e "${YELLOW}🚢 Deploying application...${NC}"
kubectl kustomize k8s/overlays/$ENVIRONMENT | kubectl apply -f -

# Deploy Gateway API resources
echo -e "${YELLOW}🌐 Deploying Gateway API resources...${NC}"
kubectl apply -f k8s/gateway-api/

# Wait for deployments to be ready
echo -e "${YELLOW}⏳ Waiting for deployments to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s \
    deployment/management-api \
    deployment/gateway-controlplane \
    deployment/auth-api \
    deployment/tenant-admin \
    deployment/backoffice \
    --namespace=tridorian-ztna

# Check status
echo ""
echo -e "${GREEN}✅ Deployment complete!${NC}"
echo ""
echo -e "${YELLOW}📊 Deployment status:${NC}"
kubectl get pods -n tridorian-ztna
echo ""
kubectl get svc -n tridorian-ztna
echo ""
kubectl get gateway -n tridorian-ztna
echo ""

# Get Gateway IP
echo -e "${YELLOW}🌍 Gateway IP Address:${NC}"
kubectl get gateway tridorian-ztna-gateway -n tridorian-ztna -o jsonpath='{.status.addresses[0].value}'
echo ""
echo ""

echo -e "${GREEN}🎉 Deployment successful!${NC}"
echo ""
echo "Next steps:"
echo "1. Update DNS records to point to the Gateway IP"
echo "2. Verify SSL certificates are provisioned"
echo "3. Test the endpoints"
echo ""
