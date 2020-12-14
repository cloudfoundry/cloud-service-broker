#!/usr/bin/env bash

set -o nounset

wait_for_service() {
    local SERVICE_NAME=$1
    local OPERATION_IN_PROGRESS=$2
    while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}" >/dev/null; do
        sleep 5
    done

    echo $#

    local LOCAL_RESULT=0
    if [ $# -gt 2 ]; then
        LOCAL_RESULT=1
        if cf service "${SERVICE_NAME}" | grep "$3" >/dev/null; then
            LOCAL_RESULT=0
        fi
    fi

    return ${LOCAL_RESULT}
}

delete_service() {
    local SERVICE_NAME=$1
    shift
    local LOCAL_RESULT=1
    for RETRY in 0 1; do
        if cf delete-service -f "${SERVICE_NAME}"; then
            wait_for_service "${SERVICE_NAME}" "delete in progress"
            if cf service "${SERVICE_NAME}" | grep "delete failed" >/dev/null; then
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
    local SERVICE_NAME=$1
    shift
    local PLAN=$1
    shift
    local NAME=$1
    shift
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
            LOCAL_RESULT=$?
            echo "Failed creating ${NAME}: ${LOCAL_RESULT}"
            cf service "${NAME}"
            delete_service "${NAME}"
        fi
    fi

    return ${LOCAL_RESULT}
}

update_service_plan() {
    local INSTANCE_NAME=$1
    shift
    local PLAN=$1
    shift

    if [ $# -gt 0 ]; then
        cf update-service "${INSTANCE_NAME}" -p "${PLAN}" -c "$@"
    else
        cf update-service "${INSTANCE_NAME}" -p "${PLAN}"
    fi

    local LOCAL_RESULT=$?
    if [ ${LOCAL_RESULT} -eq 0 ]; then
        if wait_for_service "${INSTANCE_NAME}" "update in progress" "update succeeded"; then
            echo "Successfully updated ${INSTANCE_NAME}"
        else
            cf service "${INSTANCE_NAME}"
            LOCAL_RESULT=1
        fi
    fi

    return ${LOCAL_RESULT}
}

update_service_params() {
    local INSTANCE_NAME=$1
    shift

    cf update-service "${INSTANCE_NAME}" -c "$@"

    local LOCAL_RESULT=$?
    if [ ${LOCAL_RESULT} -eq 0 ]; then
        if wait_for_service "${INSTANCE_NAME}" "update in progress" "update succeeded"; then
            echo "Successfully updated ${INSTANCE_NAME}"
        else
            cf service "${INSTANCE_NAME}"
            LOCAL_RESULT=1
        fi
    fi

    return ${LOCAL_RESULT}
}

in_list() {
    local search="$1"
    shift
    local list=("$@")
    for item in "${list[@]}"; do
        echo $item
        [[ $item == $search ]] && return 0
    done
    return 1
}

bind_service_test() {
    APP=$1
    SERVICE_INSTANCE_NAME=$2
    LOCAL_RESULT=0
    if cf bind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
        if cf restart "${APP}"; then
            sleep 10
            if cf app "${APP}" | grep running; then
                echo "successfully bound and restarted ${APP}"
            else
                LOCAL_RESULT=$?
                echo "Failed to restart ${APP}: ${LOCAL_RESULT}"
                cf env "${APP}"
                cf logs "${APP}" --recent
            fi
        else
            LOCAL_RESULT=$?
            echo "Failed to restart ${APP}: ${LOCAL_RESULT}"
            cf env "${APP}"
            cf logs "${APP}" --recent
        fi
    else
        LOCAL_RESULT=$?
        echo "Failed to bind-service ${APP} to ${SERVICE_INSTANCE_NAME}: ${LOCAL_RESULT}"
    fi
    cf stop "${APP}"
    if cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
        echo "successfully bound and restarted ${APP}"
    else
        sleep 10
        if cf unbind-service "${APP}" "${SERVICE_INSTANCE_NAME}"; then
            LOCAL_RESULT=0
        else
            LOCAL_RESULT=$?
            echo "failed to unbind-service ${APP} ${SERVICE_INSTANCE_NAME} after 2 tries"
        fi
    fi

    return ${LOCAL_RESULT}
}
