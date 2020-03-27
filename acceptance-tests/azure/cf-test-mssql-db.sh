#!/usr/bin/env bash

set -o nounset

wait_for_service() {
    SERVICE_NAME=$1
    OPERATION_IN_PROGRESS=$2
    while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}"; do
        sleep 30
    done  
}

MSSQL_SERVER_INSTANCE_NAME="test-mssql-server-$$"
cf create-service csb-azure-mssql-server standard "${MSSQL_SERVER_INSTANCE_NAME}"

wait_for_service "${MSSQL_SERVER_INSTANCE_NAME}" "create in progress"

RESULT=0
if cf bind-service spring-music "${MSSQL_SERVER_INSTANCE_NAME}"; then
    USERNAME=$(cf env spring-music | grep '"databaseLogin":' | sed -e 's/".*": "\(.*\)",/\1/;s/^[[:blank:]]*//')
    PASSWORD=$(cf env spring-music | grep '"databaseLoginPassword":' | sed -e 's/".*": "\(.*\)",/\1/;s/^[[:blank:]]*//')
    SERVER_NAME=$(cf env spring-music | grep '"sqlServerName":' | sed -e 's/".*": "\(.*\)",/\1/;s/^[[:blank:]]*//')
    SERVER_RG=$(cf env spring-music | grep '"sqldbResourceGroup"' | sed -e 's/".*": "\(.*\)",/\1/;s/^[[:blank:]]*//')

    CONFIG="{ \
    \"server_name\":\"${SERVER_NAME}\", \
    \"server_admin_username\":\"${USERNAME}\", \
    \"server_admin_password\":\"${PASSWORD}\", \
    \"server_resource_group\":\"${SERVER_RG}\" \
    }"

    echo $CONFIG

    ../cf-test-spring-music.sh csb-azure-mssql-db medium "${CONFIG}"
    RESULT=$?

    cf unbind-service spring-music "${MSSQL_SERVER_INSTANCE_NAME}"
else
  RESULT=$?
fi

cf delete-service -f "${MSSQL_SERVER_INSTANCE_NAME}"
wait_for_service "${MSSQL_SERVER_INSTANCE_NAME}" "delete in progress"

exit ${RESULT}
