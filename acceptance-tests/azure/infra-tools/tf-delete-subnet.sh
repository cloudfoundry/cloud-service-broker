#!/usr/bin/env bash

set -o nounset
set -o errexit

if [ $# -lt 3 ]; then
    echo "Usage: $0 <resource group name> <network name> <subnet name>"
    exit 1
fi

RESOURCE_GROUP=$1; shift
NETWORK_NAME=$1; shift
NAME=$1; shift

terraform() {
    local WD=$1; shift
    local DOCKER_COMMON="--rm -v ${WD}:/terraform -w /terraform -i"
    docker run $DOCKER_COMMON \
        -e ARM_SUBSCRIPTION_ID -e ARM_CLIENT_SECRET -e ARM_TENANT_ID -e ARM_CLIENT_ID \
        -t hashicorp/terraform $@
}

TMP_DIR="${HOME}/.tf-infra-tools/${RESOURCE_GROUP}-${NETWORK_NAME}-${NAME}"

#SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

terraform ${TMP_DIR} destroy -auto-approve
