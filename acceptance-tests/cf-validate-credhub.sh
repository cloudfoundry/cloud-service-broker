#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

if [ $# -lt 1 ]; then
    echo "Usage: ${0} <service-instance-name>"
    exit 1
fi

SERVICE_INSTANCE_NAME=$1; shift
APP=spring-music

RESULT=1
if cf bind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then 
    if cf env "${APP}" | grep "credhub-ref" > /dev/null; then
        echo "Success - found credhub-ref in binding"
        RESULT=0
    else
        echo "Error: did not find credhub-ref in binding"
        cf env "${APP}"
    fi
fi

cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"

exit ${RESULT}
