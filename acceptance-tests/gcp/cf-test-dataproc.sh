#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

RESULT=1

SERVICE_NAME=csb-google-dataproc
SERVICE_INSTANCE_NAME="${SERVICE_NAME}-$$"
APP_NAME=dataproc-test-app
APP_DIR=${SCRIPT_DIR}/${APP_NAME}

if create_service "$SERVICE_NAME" standard "${SERVICE_INSTANCE_NAME}"; then
    (cd "${APP_DIR}" && cf push --no-start)
    if cf bind-service ${APP_NAME} ${SERVICE_INSTANCE_NAME}; then
        if cf start ${APP_NAME}; then
            RESULT=$?
        else
            echo "${APP_NAME} failed"
            cf logs ${APP_NAME} --recent
        fi
        cf unbind-service ${APP_NAME} ${SERVICE_INSTANCE_NAME}
        cf delete -f ${APP_NAME}
    fi
    delete_service "${SERVICE_INSTANCE_NAME}"
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCEEDED"
else
    echo "$0 FAILED"
fi

exit ${RESULT}