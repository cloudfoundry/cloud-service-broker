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

SERVICE=$1
PLAN=$2
PARAMS="{\"db_name\":\"musicdb\", \"collection_name\":\"album\", \"shard_key\":\"_id\" }"



NAME="${SERVICE}-${PLAN}-$$"

set -o nounset

cf create-service "${SERVICE}" "${PLAN}" "${NAME}" -c "${PARAMS}"

cf service "${NAME}" | grep "create in progress"

set +e
while [ $? -eq 0 ]; do
    sleep 15
    cf service "${NAME}" | grep "create in progress"
done

APP=spring-music

RESULT=0
if cf bind-service "${APP}" "${NAME}"; then
  if cf restart "${APP}"; then
    echo "successfully bound and restarted ${APP}"
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
    sleep 15
done

exit ${RESULT}
