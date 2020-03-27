#!/usr/bin/env bash

wait_for_service() {
  SERVICE_NAME=$1
  OPERATION_IN_PROGRESS=$2
  while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}"; do
    sleep 30
  done

  LOCAL_RESULT=0
  if [[ -n $3 ]]; then
    LOCAL_RESULT=1
    if cf service "${SERVICE_NAME}" | grep "$3"; then
      LOCAL_RESULT=0
    fi
  fi

  return ${LOCAL_RESULT}
}

set -o errexit
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

SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$"

if [ -z "$1" ]; then
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" || exit 0
else
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" -c "$@" || exit 0
fi

RESULT=1
if wait_for_service "${SERVICE_INSTANCE_NAME}" "create in progress" "create failed"; then
  RESULT=0
else
  echo "Create service ${SERVICE_INSTANCE_NAME} should have failed."
  cf service "${SERVICE_INSTANCE_NAME}"
fi

cf delete-service -f "${SERVICE_INSTANCE_NAME}"

wait_for_service "${SERVICE_INSTANCE_NAME}" "delete in progress"

exit ${RESULT}
