#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

azure() {
cat > "${SCRIPT_DIR}/config-files/broker-config.yaml" <<EOF
brokerpak:
  builtin:
    path: /azure-brokerpak
azure:
  tenant_id: ${ARM_TENANT_ID}
  subscription_id: ${ARM_SUBSCRIPTION_ID}
  client_id: ${ARM_CLIENT_ID}
  client_secret: ${ARM_CLIENT_SECRET}
EOF
}

gcp () {
cat > "${SCRIPT_DIR}/config-files/broker-config.yaml" <<EOF
brokerpak:
  builtin:
    path: /gcp-brokerpak
google:
  credentials: ${GOOGLE_CREDENTIALS}
  project: ${GOOGLE_PROJECT}
EOF
}

aws() {
cat > "${SCRIPT_DIR}/config-files/broker-config.yaml" <<EOF
brokerpak:
  builtin:
    path: /aws-brokerpak
aws:
  access_key: ${AWS_ACCESS_KEY_ID}
  secret_access_key: ${AWS_SECRET_ACCESS_KEY}
EOF
}

if [ $# -lt 1 ]; then
  echo "Usage: $0 <gcp|azure|aws>"
  exit 1
fi

case $1 in
  aws)
    aws
  ;;
  gcp)
    gcp
  ;;
  azure)
    azure
  ;;
  *)
  echo "unknown IaaS: valid values are 'gcp', 'azure', 'aws'"
  exit 1
esac