#!/usr/bin/env bash

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

NAME="${SERVICE}-${PLAN}-$$"

if [ -z "$1" ]; then
  cf create-service "${SERVICE}" "${PLAN}" "${NAME}"
else
  cf create-service "${SERVICE}" "${PLAN}" "${NAME}" -c "$@"
fi

set -o nounset

set +e
while cf service "${NAME}" | grep "create in progress"; do
    sleep 30
done

APP=spring-music

RESULT=0
if cf bind-service "${APP}" "${NAME}"; then
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
  echo "Failed to bind-service ${APP} to ${NAME}: ${RESULT}"
fi  

if cf unbind-service "${APP}" "${NAME}"; then
  echo "successfully bound and restarted ${APP}"
else
  RESULT=$?
  echo "failed to unbind-service ${APP} ${NAME}"
fi

cf delete-service -f "${NAME}"

while cf service "${NAME}" | grep "delete in progress"; do
    sleep 30
done

exit ${RESULT}
