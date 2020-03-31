#!/bin/bash -ex
# Release script used to deploy manually from local (default is set to staging)
# Example: RELEASE_VERSION=test SERVICE_NAME=srsaysnc SERVICE_ENV=staging PROJECT=teamko619 ./deploy/scripts/manual_release.sh

export ROOT=$(git rev-parse --show-toplevel)
export WORK_DIR=${WORK_DIR:-$ROOT/applications/caterpie}
cd "$WORK_DIR"

# Publish common environment variables
source ./deploy/scripts/common.sh

# Build the project
source ./deploy/scripts/build.sh

# Push docker image
source ./deploy/scripts/docker_push.sh

# Generates deployment manifests
source ./deploy/scripts/generate_manifests.sh

# Deploy the service
./deploy/scripts/deploy.sh
