#!/bin/bash -ex

export WORK_DIR=${WORK_DIR:-$(git rev-parse --show-toplevel)/applications/caterpie}
export PROJECT=${PROJECT:-teamko619}
export SERVICE_NAME=${SERVICE_NAME:-caterpie-dev}
export SERVICE_ENV=${SERVICE_ENV:-staging}

export CONTAINER_REGISTRY=gcr.io
export SIDECAR_CONTAINER_REGISTRY=$CONTAINER_REGISTRY/api-gateway-dev
export STATSD_CONTAINER_REGISTRY=$CONTAINER_REGISTRY/api-gateway-dev
export SECURITY_CONTAINER_REGISTRY=$CONTAINER_REGISTRY/security-mesh-images
export SERVICE_IMAGE_TAG=${RELEASE_VERSION:-v2.24}
export SERVICE_IMAGE=$CONTAINER_REGISTRY/$PROJECT/caterpie:$SERVICE_IMAGE_TAG
export PROVIDER=gcp
export REGION=us-central1