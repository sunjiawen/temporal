#!/bin/bash -ex
# Release script used to deploy single service using Spinnaker.
# The script is supposed to be invoked from the ads-reporting top
# level jenkins_release.sh which is triggered by Jenkins

[[ -n $RELEASE_VERSION ]] || { echo "RELEASE_VERSION is not set"; exit 1; }

# Publish common environment variables
source ./deploy/scripts/common.sh

# Push docker image
source ./deploy/scripts/docker_push.sh

# Generates deployment manifests
source ./deploy/scripts/generate_manifests.sh

# Deploy the service by kicking off Spinnaker
./deploy/scripts/trigger_spinnaker.sh
