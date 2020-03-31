#!/bin/bash -ex

export PROJECT=snap-temporal
export CLUSTER_NAME=snap-temporal-alpha
export INSTANCE_TAGS=snap-temporal-alpha
export CLUSTER_VERSION=latest
export INSTANCE_TYPE=e2-highmem-16
export NETWORK_NAME=prod-backend-us-central1
export SUBNET_NAME=prod-backend-us-central1-small30-subnet1
export PODS_RANGE_NAME=prod-backend-us-central1-small30-subnet1-pods
export SVCS_RANGE_NAME=prod-backend-us-central1-small30-subnet1-svcs-snap-temporal
export SERVICE_ACCOUNT=snap-temporal-gke@snap-temporal.iam.gserviceaccount.com

gcloud beta --project $PROJECT container clusters create $CLUSTER_NAME \
       --default-max-pods-per-node 8 \
       --tags $INSTANCE_TAGS \
       --cluster-version $CLUSTER_VERSION \
       --image-type UBUNTU \
       --machine-type $INSTANCE_TYPE \
       --disk-size 100 \
       --enable-stackdriver-kubernetes \
       --enable-network-policy \
       --enable-cloud-logging \
       --enable-cloud-monitoring \
       --enable-ip-alias \
       --network projects/network-xpn/global/networks/$NETWORK_NAME \
       --subnetwork projects/network-xpn/regions/us-central1/subnetworks/$SUBNET_NAME \
       --cluster-secondary-range-name $PODS_RANGE_NAME \
       --services-secondary-range-name $SVCS_RANGE_NAME \
       --num-nodes 2 \
       --region us-central1 \
       --node-locations us-central1-b,us-central1-c,us-central1-f \
       --enable-autoscaling --min-nodes 2 --max-nodes 12 \
       --addons HorizontalPodAutoscaling \
       --no-enable-autorepair \
       --service-account $SERVICE_ACCOUNT \
       --metadata=disable-legacy-endpoints=false
