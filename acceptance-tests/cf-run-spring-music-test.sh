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
    echo "Failed to bind-service ${APP} to ${SERVICE_INSTANCE_NAME}: ${LOCAL_RESULT}"
  fi  
  cf stop "${APP}"
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
    echo "Usage: ${0} <service-instance-name> [update plan]"
    exit 1
fi

SERVICE_INSTANCE_NAME=$1; shift
RESULT=1

if bind_service_test spring-music "${SERVICE_INSTANCE_NAME}"; then
    ( cd "${SCRIPT_DIR}/spring-music-validator" && cf push --no-start --no-route)
    if [ $# -gt 0 ]; then
      PLAN=$1; shift
      if update_service_plan "${SERVICE_INSTANCE_NAME}" "${PLAN}" "$@"; then
        bind_service_test spring-music-validator "${SERVICE_INSTANCE_NAME}"
        RESULT=$?
      else
        echo "$0 service plan update failed"
      fi
    else
      bind_service_test spring-music-validator "${SERVICE_INSTANCE_NAME}"
      RESULT=$?
    fi

    cf delete -f spring-music-validator
fi

if [ ${RESULT} -eq 0 ]; then
  echo "$0 SUCCEEDED"
else
  echo "$0 FAILED"
fi

exit ${RESULT}
