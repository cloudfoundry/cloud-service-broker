#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/functions.sh"

if [ $# -lt 2 ]; then
    echo "Usage: $0 <service name> <plan name> [-u update-plan]"
    exit 1
fi

SERVICE=$1; shift
PLAN=$1; shift
UPDATE_PLAN=$PLAN

if [ $# -gt 0 ] && [ $1 == '-u' ]; then
    UPDATE_PLAN=$2
    shift; shift
fi

SERVICE_INSTANCE_NAME="${SERVICE}-${PLAN}-$$"

if [ $# -eq 0 ]; then
    cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}"
else
    cf create-service "${SERVICE}" "${PLAN}" "${SERVICE_INSTANCE_NAME}" -c "$@"
fi

RESULT=0
if wait_for_service "${SERVICE_INSTANCE_NAME}" "create in progress" "create succeeded"; then

    if [ $PLAN == $UPDATE_PLAN ]; then
        "${SCRIPT_DIR}/cf-run-spring-music-test.sh" ${SERVICE_INSTANCE_NAME}
    else
        "${SCRIPT_DIR}/cf-run-spring-music-test.sh" ${SERVICE_INSTANCE_NAME} ${UPDATE_PLAN}
    fi
    RESULT=$?
else
    echo "Failed creating ${SERVICE_INSTANCE_NAME}"
    cf service "${SERVICE_INSTANCE_NAME}"
fi

delete_service "${SERVICE_INSTANCE_NAME}"

exit ${RESULT}
