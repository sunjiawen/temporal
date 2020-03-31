#!/bin/bash -ex
export SERVICE_NAME=caterpie-dev
export SERVICE_ENV=staging
export PROJECT=teamko619

export ROOT=$(git rev-parse --show-toplevel)
export WORK_DIR=${WORK_DIR:-$ROOT/applications/caterpie}
cd "$WORK_DIR"

./deploy/scripts/jenkins_release.sh