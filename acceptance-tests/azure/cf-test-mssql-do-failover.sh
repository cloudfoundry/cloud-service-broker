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
            delete_service ${SERVICE}
        fi
    fi

    return ${LOCAL_RESULT}
}

PRIMARY_SERVER_NAME=mssql-server-p-$$
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~
SERVER_RG=csb-acceptance-test-rg

CONFIG="{ \
    \"instance_name\":\"${PRIMARY_SERVER_NAME}\", \
    \"admin_username\":\"${USERNAME}\", \
    \"admin_password\":\"${PASSWORD}\", \
    \"resource_group\":\"${SERVER_RG}\", \
    \"location\":\"westus\" \
    }"

RESULT=1
MSSQL_SERVER_INSTANCE_NAME=${PRIMARY_SERVER_NAME}
if cf create-service csb-azure-mssql-server standard "${MSSQL_SERVER_INSTANCE_NAME}" -c "${CONFIG}"; then
    SECONDARY_SERVER_NAME=mssql-server-s-$$

    CONFIG="{ \
        \"instance_name\":\"${SECONDARY_SERVER_NAME}\", \
        \"admin_username\":\"${USERNAME}\", \
        \"admin_password\":\"${PASSWORD}\", \
        \"resource_group\":\"${SERVER_RG}\", \
        \"location\":\"eastus\" \
        }"

    MSSQL_SERVER_INSTANCE_NAME=${SECONDARY_SERVER_NAME}
    if cf create-service csb-azure-mssql-server standard "${MSSQL_SERVER_INSTANCE_NAME}" -c "${CONFIG}"; then

        if wait_for_service "${PRIMARY_SERVER_NAME}" "create in progress" "create succeeded"; then

            if wait_for_service "${SECONDARY_SERVER_NAME}" "create in progress" "create succeeded"; then
                FOG_NAME=mssql-server-fog-$$
                DB_NAME=fog_db-$$
                CONFIG="{
                  \"instance_name\":\"${FOG_NAME}\", \
                  \"db_name\":\"${DB_NAME}\", \
                  \"server_pair\":\"test\", \
                  \"server_credential_pairs\":{ \
                    \"test\":{ \
                      \"admin_username\":\"${USERNAME}\", \
                      \"admin_password\":\"${PASSWORD}\", \
                      \"primary\":{ \
                        \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                        \"resource_group\":\"${SERVER_RG}\" \
                      }, \
                      \"secondary\":{ \
                        \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                        \"resource_group\":\"${SERVER_RG}\" \
                      } \
                    } \
                  } \
                }"

                if create_service csb-azure-mssql-db-failover-group medium "${FOG_NAME}" "${CONFIG}"; then
                    FAILIT_NAME=failit-$$
                    if create_service csb-azure-mssql-do-failover standard "${FAILIT_NAME}" "{\"fog_name\":\"${FOG_NAME}\", \"fog_resource_group\":\"${SERVER_RG}\", \"db_name\":\"${DB_NAME}\" }"; then
                        echo
                        echo Failover occurred successfully.
                        echo 
                        sleep 120
                        delete_service "${FAILIT_NAME}"
                    fi

                    delete_service "${FOG_NAME}"
                fi
            fi
        fi
    fi
    cf delete-service -f "${SECONDARY_SERVER_NAME}"
    cf delete-service -f "${PRIMARY_SERVER_NAME}"
    wait_for_service "${PRIMARY_SERVER_NAME}" "delete in progress" 
    wait_for_service "${SECONDARY_SERVER_NAME}" "delete in progress" 
else
    cf delete-service -f "${PRIMARY_SERVER_NAME}"
    wait_for_service "${PRIMARY_SERVER_NAME}" "delete in progress" 
fi

exit ${RESULT}                