#!/bin/bash
PROJECT_ID="trivpn-demo-prj"
REPO_NAME="triztna"
IMAGE_TAG="latest"
REGISTRY="asia-southeast1-docker.pkg.dev"

SERVICES=("management-api" "auth-api" "gateway-controlplane")

for SERVICE in "${SERVICES[@]}"; do
    echo "🔨 Building $SERVICE..."
    ln -sf "Dockerfile.$SERVICE" Dockerfile
    gcloud builds submit --tag "$REGISTRY/$PROJECT_ID/$REPO_NAME/$SERVICE:$IMAGE_TAG" .
done

rm Dockerfile
echo "✅ All builds complete!"
