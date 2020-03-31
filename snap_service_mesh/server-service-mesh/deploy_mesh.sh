#!/bin/bash -eu

# Follow onboarding steps here: https://wiki.sc-corp.net/display/INF/GCP+Onboarding+for+API+Gateway+and+Service+Mesh
# Deploy
#
# Default configs
#
if [ -z "$SERVICE_NAME" ]; then
    echo "No SERVICE_NAME environment variable found, please provide one"
    exit 1
fi

if [ -z "$REGION" ]; then
    echo "No REGION environment variable found, please provide one"
    exit 1
fi

if [ -z "$PROVIDER" ]; then
    echo "No PROVIDER environment variable found, please use deploy_aws.sh or deploy_gcp.sh"
    exit 1
fi

export STAGE="${STAGE:-PROD}"

# The XDS cluster matching Switchboard configuration. This is used for formatting XDS_CLUSTER_NAME below
export SWITCHBOARD_CLUSTER="${SWITCHBOARD_CLUSTER:-default}"
# if the cluster is running in stage PROD, the stage does not need to be appended to XDS_CLUSTER_NAME
if [ $(echo "${STAGE}" | tr "[:lower:]" "[:upper:]") == PROD ]; then
    export XDS_CLUSTER_NAME="cluster.$SERVICE_NAME.$SWITCHBOARD_CLUSTER.$PROVIDER-$REGION"
else
    export XDS_CLUSTER_NAME="cluster.$SERVICE_NAME.$SWITCHBOARD_CLUSTER.$PROVIDER-$REGION.$(echo "${STAGE}" | tr "[:lower:]" "[:upper:]")"
fi

if [ -z "$CONTAINER_TEMPLATE" ]; then
    echo "No CONTAINER_TEMPLATE environment variable found, defaulting to grpc-echo sample"
    CONTAINER_TEMPLATE="../grpc-echo/service-container.tmpl.yaml"
fi

if [ ! -f "$CONTAINER_TEMPLATE" ]; then
    echo "CONTAINER_TEMPLATE value $CONTAINER_TEMPLATE does not point to a file"
    exit 1
fi

# StatsD is disabled by default, see https://wiki.sc-corp.net/display/INF/Metrics on how to enable it
export ENABLE_STATSD="${ENABLE_STATSD:-false}"
if [ "$ENABLE_STATSD" != "false" ] && [ -z "$METRICS_PREFIX" ]; then
    echo "Statsd enabled but no METRICS_PREFIX provided"
    exit 1
fi

# load common container versions from config/common
. ../config/load.sh --file ../config/common/000-images.yaml --override $PROVIDER

export COMMON_CONTAINER_REGISTRY="${IMAGE_REG}"

export SERVICE_YAML=$(dockerize -template $CONTAINER_TEMPLATE)

#
# Deploy the service
#

# Publish the auth-sidecar config-map before making the deployment
if dockerize -template auth-sidecar-configmap.tmpl.yaml | kubectl apply -f -; then
    echo "Successfully updated auth-sidecar config-map for $SERVICE_NAME deployment"
else
    echo "Failed to update auth-sidecar config-map for $SERVICE_NAME deployment"
    exit 1
fi

# Update the deployment
if dockerize -template deployment.tmpl.yaml | kubectl apply -f -; then
	echo "Successfully updated $SERVICE_NAME deployment"
else
	echo "Failed to update $SERVICE_NAME deployment"
	exit 1
fi

#if [ -n "$CREATE_LB" ]; then
#	# Expose the service behind LB
#	if dockerize -template temporal-server-container.tmpl.yaml | kubectl apply -f -; then
#		echo "Successfully exposed $SERVICE_NAME deployment"
#	else
#		echo "Failed to expose $SERVICE_NAME deployment"
#		exit 1
#	fi
#fi
