#!/bin/bash -ex

[[ -n $PROVIDER ]] || { echo "PROVIDER is not set"; exit 1; }
[[ -n $REGION ]] || { echo "REGION is not set"; exit 1; }
[[ -n $SERVICE_NAME ]] || { echo "SERVICE_NAME is not set"; exit 1; }
[[ -n $SERVICE_ENV ]] || { echo "SERVICE_ENV is not set"; exit 1; }

export WORK_DIR=${WORK_DIR:-$(git rev-parse --show-toplevel)/applications/caterpie}

# Container image values used by dockerize to generate manifests
export IMAGE_SIDECAR="$SIDECAR_CONTAINER_REGISTRY/sidecar:v1.12.2-463.cff6cbc2.cff6cbc2"
export IMAGE_STATSD="$STATSD_CONTAINER_REGISTRY/statsd-pubsub:v1.1.0-131.cd6b8ba5"
export IMAGE_SIDECAR_INIT="$SIDECAR_CONTAINER_REGISTRY/sidecar-init:v0.1.6"
export IMAGE_AUTH_SIDECAR="$SECURITY_CONTAINER_REGISTRY/auth-sidecar:v1.16.5"
if [ "$SERVICE_ENV" = "prod" ]; then
  export CONFIG_FILE="$WORK_DIR/deploy/configs/config_prod.properties"
  export CLUSTER_STAGE="STAGE_PROD"
  export SWITCHBOARD_CLUSTER="caterpie"
  export XDS_CLUSTER_NAME="cluster.$SERVICE_NAME.$SWITCHBOARD_CLUSTER.$PROVIDER-$REGION"
else
  export CONFIG_FILE="$WORK_DIR/deploy/configs/config_staging.properties"
  export CLUSTER_STAGE="STAGE_STAGING"
  export SWITCHBOARD_CLUSTER="caterpie-dev"
  export XDS_CLUSTER_NAME="cluster.$SERVICE_NAME.$SWITCHBOARD_CLUSTER.$PROVIDER-$REGION.STAGING"
fi

# Generated manifiests used by subsequent deployment step
export MANIFESTS_DIR="generated_manifests/$SERVICE_ENV"
export AUTH_CONFIGMAP_YAML="$MANIFESTS_DIR/auth-sidecar-configmap.yaml"
export SERVICE_CONFIGMAP_YAML="$MANIFESTS_DIR/$SERVICE_NAME-configmap.yaml"
export SERVICE_DEPLOYMENT_YAML="$MANIFESTS_DIR/$SERVICE_NAME-deployment.yaml"

rm -rf "$MANIFESTS_DIR" && mkdir -p "$MANIFESTS_DIR"
dockerize -template "$WORK_DIR/deploy/templates/auth-sidecar-configmap.tmpl.yaml" > "$AUTH_CONFIGMAP_YAML"
dockerize -template "$WORK_DIR/deploy/templates/deployment.tmpl.yaml" > "$SERVICE_DEPLOYMENT_YAML"
kubectl create configmap caterpie-configmap \
    --from-file=config.properties="$CONFIG_FILE" -o yaml \
    --dry-run > "$SERVICE_CONFIGMAP_YAML"

echo "Successfully generated all manifests for $SERVICE_NAME"
