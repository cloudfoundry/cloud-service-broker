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

PRIMARY_SERVER_NAME=mssql-server-p-$$
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~
SERVER_RG=csb-acceptance-test-rg

CONFIG="{ \
    \"instance_name\":\"${PRIMARY_SERVER_NAME}\", \
    \"admin_username\":\"${USERNAME}\", \
    \"admin_password\":\"${PASSWORD}\", \
    \"resource_group\":\"${SERVER_RG}\", \
    \"region\":\"westus\" \
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
        \"region\":\"eastus\" \
        }"

    MSSQL_SERVER_INSTANCE_NAME=${SECONDARY_SERVER_NAME}
    if cf create-service csb-azure-mssql-server standard "${MSSQL_SERVER_INSTANCE_NAME}" -c "${CONFIG}"; then

        if wait_for_service "${PRIMARY_SERVER_NAME}" "create in progress" "create succeeded"; then

            if wait_for_service "${SECONDARY_SERVER_NAME}" "create in progress" "create succeeded"; then
                FOG_NAME=mssql-server-fog-$$
                CONFIG="{
                  \"instance_name\":\"${FOG_NAME}\", \
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

                echo $CONFIG

                ../cf-test-spring-music.sh csb-azure-mssql-db-failover-group medium "${CONFIG}"
                RESULT=$?
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