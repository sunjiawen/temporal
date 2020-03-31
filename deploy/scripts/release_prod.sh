#!/bin/bash -ex
export SERVICE_NAME=async-srs-prod
export SERVICE_ENV=prod
export PROJECT=adsapisc

export ROOT=$(git rev-parse --show-toplevel)
export WORK_DIR=${WORK_DIR:-$ROOT/applications/caterpie}
cd "$WORK_DIR"

./deploy/scripts/jenkins_release.sh