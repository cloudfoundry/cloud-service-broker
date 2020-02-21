#!/usr/bin/env bash

if [[ -z ${env} ]]; then
  echo 'Missing env variable pointing to smith env file'
  exit 1
fi

export PCF_NETWORK=$(cat $env | jq -r .service_network_name)
export REGION=$(cat $env | jq -r .region)
export PROJECT=$(cat $env | jq -r .project)

export NETWORK="https://www.googleapis.com/compute/alpha/projects/${PROJECT}/global/networks/${PCF_NETWORK}"
gcloud beta sql instances create csb-mysql-1 --network $NETWORK --tier db-n1-standard-1 --region ${REGION}  --no-assign-ip