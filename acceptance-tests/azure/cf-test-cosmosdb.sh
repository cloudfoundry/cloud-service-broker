#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

SERVICE=csb-azure-cosmosdb-sql
SERVICE_NAME=cosmosdb-sql-$$
PLAN=medium
APP_NAME=cosmosdb-test-app
APP_DIR=${SCRIPT_DIR}/${APP_NAME}

RESULT=1
if create_service ${SERVICE} ${PLAN} ${SERVICE_NAME}; then
    (cd "${APP_DIR}" && cf push --no-start)
    if cf bind-service ${APP_NAME} ${SERVICE_NAME}; then
        if cf start ${APP_NAME}; then
            RESULT=0
            echo "${APP_NAME} success"
        else
            echo "${APP_NAME} failed"
            cf logs ${APP_NAME} --recent
        fi
        cf delete -f ${APP_NAME}
    fi
    delete_service ${SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi

exit ${RESULT}
