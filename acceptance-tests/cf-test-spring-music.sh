#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

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

SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$"

if [ -z "$1" ]; then
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}"
else
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" -c "$@"
fi

RESULT=0
if wait_for_service "${SERVICE_INSTANCE_NAME}" "create in progress" "create succeeded"; then

  "${SCRIPT_DIR}"/cf-run-spring-music-test.sh ${SERVICE_INSTANCE_NAME}
  RESULT=$?
else
  echo "Failed creating ${SERVICE_INSTANCE_NAME}"
  cf service "${SERVICE_INSTANCE_NAME}"
fi

delete_service "${SERVICE_INSTANCE_NAME}"

exit ${RESULT}
