#!/usr/bin/env bash

set -o nounset
set -o errexit

#set -o xtrace

if [ $# -lt 3 ]; then
    echo "Usage: $0 <vpc id> <region> <subnet name>"
    exit 1
fi

VPC_ID=$1; shift
REGION=$1; shift
NAME=$1; shift

terraform() {
    local WD=$1; shift
    local DOCKER_COMMON="--rm -v ${WD}:/terraform -w /terraform -i"
    docker run $DOCKER_COMMON \
        -e AWS_SECRET_ACCESS_KEY -e AWS_ACCESS_KEY_ID \
        -t hashicorp/terraform $@
}

TMP_DIR="${HOME}/.tf-infra-tools/rds-sg-${VPC_ID}-${NAME}"
mkdir -p "${TMP_DIR}"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cp ${SCRIPT_DIR}/terraform/rds-subnet-group.tf ${TMP_DIR}

cat >${TMP_DIR}/terraform.tfvars <<EOL
name = "${NAME}"
vpc_id = "${VPC_ID}" 
region = "${REGION}"
EOL

terraform ${TMP_DIR} init > /dev/null
terraform ${TMP_DIR} apply -auto-approve -no-color > /dev/null
RDS_SUBNET_GROUP=$(terraform ${TMP_DIR} output -json name | jq -r .)
VPC_SECURITY_GROUP_ID=$(terraform ${TMP_DIR} output -json security_group_id | jq -r .)

echo "{\"rds_subnet_group\" : \"${RDS_SUBNET_GROUP}\", \"rds_vpc_security_group_ids\" : \"${VPC_SECURITY_GROUP_ID}\"}"

