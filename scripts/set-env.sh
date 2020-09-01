#!/usr/bin/env bash

set +x # Hide secrets

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "You must source this script\nsource ${0}" && exit 1

export ARM_SUBSCRIPTION_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .subscription_id)
export ARM_TENANT_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .tenant_id)
export ARM_CLIENT_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .client_id)
export ARM_CLIENT_SECRET=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/Azure Service Account Key" | jq -r .client_secret)

export GCP_SERVICE_ACCOUNT_JSON=$(lpass show --notes "Shared-CF Platform Engineering/pks cluster management gcp service account")
export ROOT_SERVICE_ACCOUNT_JSON="${GCP_SERVICE_ACCOUNT_JSON}"
export GOOGLE_CREDENTIALS="${GCP_SERVICE_ACCOUNT_JSON}"
export GOOGLE_PROJECT=$(echo ${GOOGLE_CREDENTIALS} | jq -r .project_id)

export GCP_PAS_NETWORK=$(lpass show "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml" | grep gcp-network | cut -d ' ' -f 2)

export AWS_ACCESS_KEY_ID=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service AWS credentials" | jq -r .access_key)
export AWS_SECRET_ACCESS_KEY=$(lpass show --notes "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service AWS credentials" | jq -r .secret_key)

export AWS_PAS_VPC_ID=$(lpass show "Shared-CF Platform Engineering/pe-cloud-service-broker/cloud service broker pipeline secrets.yml" | grep aws-vpc-id | cut -d ' ' -f 2)

export SECURITY_USER_NAME=brokeruser
export SECURITY_USER_PASSWORD=brokeruserpassword
export DB_HOST=localhost
export DB_USERNAME=broker
export DB_PASSWORD=brokerpass
export DB_NAME=brokerdb
export PORT=8080