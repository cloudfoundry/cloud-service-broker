#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

export GSB_API_USER='csb-broker'
export GSB_API_PASSWORD='S0m3S3cr3tPa$$w0rd'
export PORT=8080
export GSB_API_HOSTNAME=$(kubectl get service csb --output='jsonpath="{.status.loadBalancer.ingress[0].ip}"' | tr -d '"')

${SCRIPT_DIR}/../build/cloud-service-broker.darwin client run-examples