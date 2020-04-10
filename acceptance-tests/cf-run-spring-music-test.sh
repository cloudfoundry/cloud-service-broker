#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

bind_service_test() {
  APP=$1
  SERVICE_INSTANCE_NAME=$2
  LOCAL_RESULT=0
  if cf bind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
    if cf restart "${APP}"; then
      sleep 10
      if cf app "${APP}" | grep running; then
        echo "successfully bound and restarted ${APP}"
      else
        LOCAL_RESULT=$?
        echo "Failed to restart ${APP}: ${RESULT}"
        cf env "${APP}"
        cf logs "${APP}" --recent
      fi
    else
      LOCAL_RESULT=$?
      echo "Failed to restart ${APP}: ${RESULT}"
      cf env "${APP}"
      cf logs "${APP}" --recent
    fi
  else
    LOCAL_RESULT=$?
    echo "Failed to bind-service ${APP} to ${SERVICE_INSTANCE_NAME}: ${RESULT}"
  fi  
  
  if cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
    echo "successfully bound and restarted ${APP}"
  else
    sleep 10
    if cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
      LOCAL_RESULT=0
    else
      LOCAL_RESULT=$?
      echo "failed to unbind-service ${APP} ${SERVICE_INSTANCE_NAME} after 2 tries"
    fi
  fi

  return ${LOCAL_RESULT}
}

if [ $# -lt 1 ]; then
    echo "Usage: ${0} <service-instance-name>"
    exit 1
fi

SERVICE_INSTANCE_NAME=$1; shift

if bind_service_test spring-music "${SERVICE_INSTANCE_NAME}"; then
    ( cd "${SCRIPT_DIR}/spring-music-validator" && cf push --no-start )

    bind_service_test spring-music-validator "${SERVICE_INSTANCE_NAME}"
    RESULT=$?

    cf delete -f spring-music-validator
fi

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}
