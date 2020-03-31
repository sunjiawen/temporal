#!/bin/bash -eu
# Script to trigger spinnaker deployment using spinnaker cli (go-spinnaker)
# Example:
#   RELEASE_VERSION=v1 \
#   SERVICE_NAME=srsaysnc \
#   AUTH_CONFIGMAP_YAML=generated_manifests/staging/auth-sidecar-configmap.yaml \
#   SERVICE_CONFIGMAP_YAML=generated_manifests/staging/srsaysnc-configmap.yaml \
#   SERVICE_DEPLOYMENT_YAML=generated_manifests/staging/srsaysnc-deployment.yaml \
#   ./deploy/scripts/trigger_spinnaker.sh

[[ -n $SERVICE_NAME ]] || { echo "SERVICE_NAME is not set"; exit 1; }
[[ -n $RELEASE_VERSION ]] || { echo "RELEASE_VERSION is not set"; exit 1; }
[[ -s $AUTH_CONFIGMAP_YAML ]] || { echo "$AUTH_CONFIGMAP_YAML is missing"; exit 1; }
[[ -s $SERVICE_CONFIGMAP_YAML ]] || { echo "$SERVICE_CONFIGMAP_YAML is missing"; exit 1; }
[[ -s $SERVICE_DEPLOYMENT_YAML ]] || { echo "$SERVICE_DEPLOYMENT_YAML is missing"; exit 1; }

if [ "$SERVICE_ENV" = "prod" ]; then
    go-spinnaker deployment publish \
        --use-convoy \
        -n asyncsrs \
        -v $(date +"%Y-%m-%dT%H:%M:%SZ") \
        -a "targetEnv"="staging","targetRegion"="us-central1","targetService"="$SERVICE_NAME" \
        -o "version"="$RELEASE_VERSION" \
        -f "$AUTH_CONFIGMAP_YAML" \
        -f "$SERVICE_CONFIGMAP_YAML" \
        -f "$SERVICE_DEPLOYMENT_YAML"
else
    go-spinnaker deployment publish \
        -t projects/spinnaker-snap-prod/topics/dev-artifacts \
        -n asyncsrs \
        -v $(date +"%Y-%m-%dT%H:%M:%SZ") \
        -a "targetEnv"="staging","targetRegion"="us-central1","targetService"="$SERVICE_NAME" \
        -o "version"="$RELEASE_VERSION" \
        -f "$AUTH_CONFIGMAP_YAML" \
        -f "$SERVICE_CONFIGMAP_YAML" \
        -f "$SERVICE_DEPLOYMENT_YAML" \
        -b snapengine-maven-publish-dev
fi