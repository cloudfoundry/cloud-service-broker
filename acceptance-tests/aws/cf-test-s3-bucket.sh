#!/usr/bin/env bash

set -o pipefail
set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

SERVICE_NAME=s3-bucket-$$

RESULT=1
if create_service csb-aws-s3-bucket private ${SERVICE_NAME}; then
    (cd "${SCRIPT_DIR}/s3-test-app" && cf push --no-start)
    if cf bind-service s3-test-app ${SERVICE_NAME}; then
        if cf start s3-test-app; then
            RESULT=0
            echo "s3-test-app success"
        else
            echo "s3-test-app failed"
            cf logs s3-test-app --recent
        fi
        cf delete -f s3-test-app
    fi
    delete_service ${SERVICE_NAME}
fi

if [ ${RESULT} -eq 0 ]; then
    echo "$0 SUCCESS"
else
    echo "$0 FAILED"
fi

exit ${RESULT}
