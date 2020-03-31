#!/bin/bash -e

export SERVICE_NAME=snap-temporal
export REGION=us-central1
export PROVIDER=gcp
export STAGE=PROD
export SWITCHBOARD_CLUSTER=snap-temporal-alpha
export CONTAINER_TEMPLATE=temporal-server-container.tmpl.yaml
export ENABLE_STATSD=true
export CREATE_LB=false
export METRICS_PREFIX=snap-temporal

# Follow onboarding steps here: https://wiki.sc-corp.net/display/INF/GCP+Onboarding+for+API+Gateway+and+Service+Mesh
# Deploy
. deploy_mesh.sh