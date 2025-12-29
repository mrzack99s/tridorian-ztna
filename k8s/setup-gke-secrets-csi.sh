#!/bin/bash

# Setup GKE Secrets Store CSI Driver
# This script enables and configures the native GKE secret management

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_ID="${GCP_PROJECT_ID:-trivpn-demo-prj}"
CLUSTER_NAME="${GKE_CLUSTER_NAME:-triztna-dev-cluster}"
REGION="${GKE_REGION:-asia-southeast1}"
NAMESPACE="tridorian-ztna"

echo -e "${GREEN}🔐 Setting up GKE Secrets Store CSI Driver${NC}"
echo "=============================================="
echo "Project: $PROJECT_ID"
echo "Cluster: $CLUSTER_NAME"
echo "Region: $REGION"
echo ""

# Create GCP Service Account for Secret Manager access
echo -e "${YELLOW}🔑 Setting up GCP Service Account...${NC}"

SA_NAME="gke-secrets-sa"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# Create service account if it doesn't exist
if ! gcloud iam service-accounts describe $SA_EMAIL --project=$PROJECT_ID &> /dev/null; then
    gcloud iam service-accounts create $SA_NAME \
        --display-name="GKE Secrets Store CSI Driver Service Account" \
        --project=$PROJECT_ID
fi

# Get cluster credentials
echo -e "${YELLOW}🔑 Getting cluster credentials...${NC}"
gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION

# Grant Secret Manager access
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="roles/secretmanager.secretAccessor"

gcloud secrets add-iam-policy-binding jwt-private-key \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding jwt-public-key \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding ztna-db-username \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding ztna-db-password \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding ztna-db-name \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding ztna-db-host \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

gcloud secrets add-iam-policy-binding ztna-cache-password \
   --role=roles/secretmanager.secretAccessor \
   --member=principal://iam.googleapis.com/projects/484715688142/locations/global/workloadIdentityPools/trivpn-demo-prj.svc.id.goog/subject/ns/tridorian-ztna/sa/gke-secrets-sa

# Enable Workload Identity
echo -e "${YELLOW}🔗 Configuring Workload Identity...${NC}"

# Create Kubernetes service account
kubectl create serviceaccount gke-secrets-sa \
    --namespace=$NAMESPACE \
    --dry-run=client -o yaml | kubectl apply -f -

# Bind GCP SA to K8s SA
gcloud iam service-accounts add-iam-policy-binding $SA_EMAIL \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/gke-secrets-sa]" \
    --project=$PROJECT_ID

# Annotate K8s service account
kubectl annotate serviceaccount gke-secrets-sa \
    --namespace=$NAMESPACE \
    iam.gke.io/gcp-service-account=$SA_EMAIL \
    --overwrite

echo ""
echo -e "${GREEN}✅ GKE Secrets Store CSI Driver setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Apply SecretProviderClass: kubectl apply -f k8s/base/secretproviderclass.yaml"
echo "2. Deploy applications (secrets will be mounted automatically)"
echo "3. Verify: kubectl get secretproviderclass -n $NAMESPACE"
echo ""
