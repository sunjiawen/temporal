#!/bin/bash -ex

[[ -s $AUTH_CONFIGMAP_YAML ]] || { echo "$AUTH_CONFIGMAP_YAML is missing"; exit 1; }
[[ -s $SERVICE_CONFIGMAP_YAML ]] || { echo "$SERVICE_CONFIGMAP_YAML is missing"; exit 1; }
[[ -s $SERVICE_DEPLOYMENT_YAML ]] || { echo "$SERVICE_DEPLOYMENT_YAML is missing"; exit 1; }
[[ -n $SERVICE_NAME ]] || { echo "SERVICE_NAME is not set"; exit 1; }
[[ -n $SERVICE_IMAGE_TAG ]] || { echo "SERVICE_IMAGE_TAG is not set"; exit 1; }
[[ -n $CONTAINER_REGISTRY ]] || { echo "CONTAINER_REGISTRY is not set"; exit 1; }
[[ -n $PROJECT ]] || { echo "PROJECT_NAME is not set"; exit 1; }
[[ -n $SWITCHBOARD_CLUSTER ]] || { echo "PROJECT_NAME is not set"; exit 1; }
[[ -n $REGION ]] || { echo "REGION is not set"; exit 1; }

#
# Deploy the service
#

gcloud config set project "$PROJECT"
gcloud container clusters get-credentials "$SWITCHBOARD_CLUSTER" --region "$REGION" --project "$PROJECT"

# Update the  auth-sidecar-configmap
if kubectl apply -f "$AUTH_CONFIGMAP_YAML"; then
    echo "Successfully updated $SERVICE_NAME auth-sidecar-configmap"
else
    echo "Failed to update $SERVICE_NAME auth-sidecar-configmap"
    exit 1
fi

# Update the caterpie-configmap
if kubectl apply -f "$SERVICE_CONFIGMAP_YAML"; then
    echo "Successfully updated $SERVICE_NAME caterpie-configmap"
else
    echo "Failed to update $SERVICE_NAME caterpie-configmap"
    exit 1
fi

# Update the deployment
if kubectl apply -f "$SERVICE_DEPLOYMENT_YAML"; then
	  kubectl annotate deployment "$SERVICE_NAME" \
	      "$CONTAINER_REGISTRY"/change-cause="Image:$SERVICE_IMAGE_TAG Date:$(date +"%D %T")" \
	      --overwrite=true
	  echo "Successfully updated $SERVICE_NAME deployment"
else
	  echo "Failed to update $SERVICE_NAME deployment"
	  exit 1
fi