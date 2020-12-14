#!/usr/bin/env bash

set -o nounset
set -o errexit

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

TMP_DIR="${HOME}/.tf-infra-tools/elasticache-sg-${VPC_ID}-${NAME}"

#SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

terraform ${TMP_DIR} destroy -auto-approve

rm -rf "${TMP_DIR}"