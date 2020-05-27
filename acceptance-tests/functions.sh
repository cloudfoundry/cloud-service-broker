#!/usr/bin/env bash

set -o nounset

wait_for_service() {
    local SERVICE_NAME=$1
    local OPERATION_IN_PROGRESS=$2
    while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}" > /dev/null; do
        sleep 5
    done

    local LOCAL_RESULT=0
    if [ $# -gt 2 ]; then
        LOCAL_RESULT=1
        if cf service "${SERVICE_NAME}" | grep "$3" > /dev/null; then
            LOCAL_RESULT=0
        fi
    fi

    return ${LOCAL_RESULT}
}

delete_service() {
    local SERVICE_NAME=$1; shift
    local LOCAL_RESULT=1
    for RETRY in 0 1; do
        if cf delete-service -f "${SERVICE_NAME}"; then
            wait_for_service "${SERVICE_NAME}" "delete in progress"
            if cf service "${SERVICE_NAME}" | grep "delete failed" > /dev/null; then
                echo "Failed to delete ${SERVICE_NAME}"
                cf service "${SERVICE_NAME}"
                LOCAL_RESULT=1
                if [ $RETRY -eq 0 ]; then
                    echo "Retry delete in 5..."
                    sleep 5
                fi
            else
                echo "Successfully deleted ${SERVICE_NAME}"
                LOCAL_RESULT=0
                break
            fi
        fi
    done
    return ${LOCAL_RESULT}
}

create_service() {
    local SERVICE_NAME=$1; shift
    local PLAN=$1; shift
    local NAME=$1; shift
    if [ $# -gt 0 ]; then
        cf create-service "${SERVICE_NAME}" "${PLAN}" "${NAME}" -c "$@"
    else
        cf create-service "${SERVICE_NAME}" "${PLAN}" "${NAME}"
    fi

    local LOCAL_RESULT=$?
    if [ ${LOCAL_RESULT} -eq 0 ]; then
        if wait_for_service "${NAME}" "create in progress" "create succeeded"; then
            echo "Successfully created ${NAME}"
        else
            cf service "${NAME}"
            delete_service "${NAME}"
            LOCAL_RESULT=1
        fi
    fi

    return ${LOCAL_RESULT}
}