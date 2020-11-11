#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

. "${SCRIPT_DIR}/functions.sh"

if [ $# -lt 1 ]; then
    echo "Usage: ${0} <service-instance-name> [update plan]"
    exit 1
fi

SERVICE_INSTANCE_NAME=$1
shift
RESULT=1

if bind_service_test spring-music "${SERVICE_INSTANCE_NAME}"; then
    (cd "${SCRIPT_DIR}/spring-music-validator" && cf push --no-start --no-route)
    if [ $# -gt 0 ]; then
        PLAN=$1
        shift
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
