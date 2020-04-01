#!/usr/bin/env bash

set -o nounset

wait_for_service() {
    SERVICE_NAME=$1
    OPERATION_IN_PROGRESS=$2
    while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}"; do
        sleep 30
    done

    LOCAL_RESULT=0
    if [ $# -gt 2 ]; then
        LOCAL_RESULT=1
        if cf service "${SERVICE_NAME}" | grep "$3"; then
            LOCAL_RESULT=0
        fi
    fi

    return ${LOCAL_RESULT}
}

delete_service() {
    NAME=$1; shift
    LOCAL_RESULT=1
    if cf delete-service -f "${NAME}"; then
        wait_for_service "${NAME}" "delete in progress"
        LOCAL_RESULT=$?
    fi

    return ${LOCAL_RESULT}
}

create_service() {
    SERVICE=$1; shift
    PLAN=$1; shift
    NAME=$1; shift
    if [ $# -gt 0 ]; then
        cf create-service "${SERVICE}" "${PLAN}" "${NAME}" -c "$@"
    else
        cf create-service "${SERVICE}" "${PLAN}" "${NAME}"
    fi

    LOCAL_RESULT=$?
    if [ ${LOCAL_RESULT} -eq 0 ]; then
        if wait_for_service "${NAME}" "create in progress" "create succeeded"; then
            echo
        else
            LOCAL_RESULT=$?
            cf service "${NAME}"
            delete_service "${NAME}"
        fi
    fi

    return ${LOCAL_RESULT}
}