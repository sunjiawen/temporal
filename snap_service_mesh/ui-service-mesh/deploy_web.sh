#!/bin/bash -e

export SERVICE_NAME=snap-cadence-ui-dev
export REGION=us-central1
export PROVIDER=gcp
export STAGE=STAGING
export SWITCHBOARD_CLUSTER=snap-cadence-dev-1
export CONTAINER_TEMPLATE=cadence-web-container.tmpl.yaml
export ENABLE_STATSD=false
export CREATE_LB=false

# Follow onboarding steps here: https://wiki.sc-corp.net/display/INF/GCP+Onboarding+for+API+Gateway+and+Service+Mesh
# Deploy
. deploy_mesh.sh