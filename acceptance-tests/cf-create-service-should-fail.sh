#!/usr/bin/env bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

set -o pipefail

if [ -z "$1" ]; then
    echo "No service name argument supplied"
    exit 1
fi

if [ -z "$2" ]; then
    echo "No plan argument supplied"
    exit 1
fi

SERVICE=$1; shift
PLAN=$1; shift

SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$-f"

create_service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" "$@"

if [ $? -ne 0 ]; then
  if cf service ${SERVICE_INSTANCE_NAME}; then 
    echo "Purging service instance..."
    cf purge-service-instance -f "${SERVICE_INSTANCE_NAME}"
  fi
  echo "$0 SUCCEEDED"
  exit 0
fi

delete_service "${SERVICE_INSTANCE_NAME}"
echo "$0 FAILED"

exit 1
