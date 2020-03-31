#!/bin/bash -ex

[[ -n $SERVICE_IMAGE ]] || { echo "SERVICE_IMAGE is not set"; exit 1; }

docker build -t "$SERVICE_IMAGE" .
docker push "$SERVICE_IMAGE"