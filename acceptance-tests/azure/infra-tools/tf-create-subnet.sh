#!/usr/bin/env bash

set -o nounset
set -o errexit

#set -o xtrace

if [ $# -lt 3 ]; then
    echo "Usage: $0 <resource group name> <network name> <subnet name> [subnet cidr - defautl 10.0.10.0/26]"
    exit 1
fi

RESOURCE_GROUP=$1; shift
NETWORK_NAME=$1; shift
NAME=$1; shift

SUBNET_CIDR="10.0.10.0/26"
if [ $# -gt 0 ]; then
    SUBNET_CIDR=$1; shift
fi

terraform() {
    local WD=$1; shift
    local DOCKER_COMMON="--rm -v ${WD}:/terraform -w /terraform -i"
    docker run $DOCKER_COMMON \
        -e ARM_SUBSCRIPTION_ID -e ARM_CLIENT_SECRET -e ARM_TENANT_ID -e ARM_CLIENT_ID \
        -t hashicorp/terraform $@
}

TMP_DIR="${HOME}/.tf-infra-tools/${RESOURCE_GROUP}-${NETWORK_NAME}-${NAME}"
mkdir -p "${TMP_DIR}"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cp ${SCRIPT_DIR}/terraform/subnet.tf ${TMP_DIR}

cat >${TMP_DIR}/terraform.tfvars <<EOL
name = "${NAME}"
resource_group_name = "${RESOURCE_GROUP}"
virtual_network_name = "${NETWORK_NAME}" 
subnet_cidr = "${SUBNET_CIDR}"
EOL

terraform ${TMP_DIR} init #> /dev/null
terraform ${TMP_DIR} apply -auto-approve #> /dev/null
terraform ${TMP_DIR} output id
