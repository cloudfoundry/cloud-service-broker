#!/usr/bin/env bash
bind_service_test() {
  APP=$1
  SERVICE_INSTANCE_NAME=$2
  RESULT=0
  if cf bind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
    if cf restart "${APP}"; then
      sleep 10
      if cf app "${APP}" | grep running; then
        echo "successfully bound and restarted ${APP}"
      else
        RESULT=$?
        echo "Failed to restart ${APP}: ${RESULT}"
        cf env "${APP}"
        cf logs "${APP}" --recent
      fi
    else
      RESULT=$?
      echo "Failed to restart ${APP}: ${RESULT}"
      cf env "${APP}"
      cf logs "${APP}" --recent
    fi
  else
    RESULT=$?
    echo "Failed to bind-service ${APP} to ${SERVICE_INSTANCE_NAME}: ${RESULT}"
  fi  
  
  if cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
    echo "successfully bound and restarted ${APP}"
  else
    RESULT=$?
    echo "failed to unbind-service ${APP} ${SERVICE_INSTANCE_NAME}"
  fi

  return ${RESULT}
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
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}"
else
  cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" -c "$@"
fi

set -o nounset

set +e
while cf service "${SERVICE_INSTANCE_NAME}" | grep "create in progress"; do
    sleep 30
done

RESULT=0
if bind_service_test spring-music "${SERVICE_INSTANCE_NAME}"; then
  export SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

  ( cd "${SCRIPT_DIR}/spring-music-validator" && cf push --no-start )

  bind_service_test spring-music-validator "${SERVICE_INSTANCE_NAME}"
  RESULT=$?

  cf delete -f spring-music-validator
else
  RESULT=$?
fi

cf delete-service -f "${SERVICE_INSTANCE_NAME}"

while cf service "${SERVICE_INSTANCE_NAME}" | grep "delete in progress"; do
    sleep 30
done

exit ${RESULT}
