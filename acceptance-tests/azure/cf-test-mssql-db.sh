#!/usr/bin/env bash

set -o nounset

wait_for_service() {
  SERVICE_NAME=$1
  OPERATION_IN_PROGRESS=$2
  while cf service "${SERVICE_NAME}" | grep "${OPERATION_IN_PROGRESS}"; do
    sleep 30
  done

  RESULT=0
  if [ $? -gt 2 ]; then
    RESULT=1
    if cf service "${SERVICE_NAME}" | grep "$3"; then
      RESULT=0
    fi
  fi

  return ${RESULT}
}

SERVER_NAME=mssql-server-$$
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~
SERVER_RG=csb-acceptance-test-rg

CONFIG="{ \
    \"instance_name\":\"${SERVER_NAME}\", \
    \"admin_username\":\"${USERNAME}\", \
    \"admin_password\":\"${PASSWORD}\", \
    \"resource_group\":\"${SERVER_RG}\" \
    }"

MSSQL_SERVER_INSTANCE_NAME="test-mssql-server-$$"

RESULT=1
if cf create-service csb-azure-mssql-server standard "${MSSQL_SERVER_INSTANCE_NAME}" -c "${CONFIG}"; then

    if wait_for_service "${MSSQL_SERVER_INSTANCE_NAME}" "create in progress" "create succeeded"; then
        CONFIG="{ \
        \"server_name\":\"${SERVER_NAME}\", \
        \"server_admin_username\":\"${USERNAME}\", \
        \"server_admin_password\":\"${PASSWORD}\", \
        \"server_resource_group\":\"${SERVER_RG}\" \
        }"

        echo $CONFIG

        ../cf-test-spring-music.sh csb-azure-mssql-db medium "${CONFIG}"
        RESULT=$?
    fi
fi
cf delete-service -f "${MSSQL_SERVER_INSTANCE_NAME}"
wait_for_service "${MSSQL_SERVER_INSTANCE_NAME}" "delete in progress"

exit ${RESULT}