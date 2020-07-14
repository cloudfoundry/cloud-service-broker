#!/usr/bin/env bash

set -o nounset

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../acceptance-tests/functions.sh"

EXT=$USER
PRIMARY_SERVER_NAME=mssql-server-p-$EXT
SECONDARY_SERVER_NAME=mssql-server-s-$EXT
USERNAME=anadminuser
PASSWORD=This_S0uld-3eC0mplex~

SERVER_RG=rg-test-service-$EXT
RESULT=1
if create_service csb-azure-resource-group standard "${SERVER_RG}" "{\"instance_name\":\"${SERVER_RG}\"}"; then
  "${SCRIPT_DIR}/../acceptance-tests/azure/cf-create-mssql-server.sh" "${PRIMARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${SERVER_RG}" westus &
  PRIMARY_PID=$!

  "${SCRIPT_DIR}/../acceptance-tests/azure/cf-create-mssql-server.sh" "${SECONDARY_SERVER_NAME}" "${USERNAME}" "${PASSWORD}" "${SERVER_RG}" eastus &
  SECONDARY_PID=$!

  if wait ${PRIMARY_PID} && wait ${SECONDARY_PID}; then
      FOG_NAME=mssql-server-fog-$EXT
      CONFIG="{
        \"instance_name\":\"${FOG_NAME}\", \
        \"db_name\":\"test_db\", \
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
          }, \
          \"test-fail\":{ \
            \"admin_username\":\"foo\", \
            \"admin_password\":\"bar\", \
            \"primary\":{ \
              \"server_name\":\"s1\", \
              \"resource_group\":\"rg\" \
            }, \
            \"secondary\":{ \
              \"server_name\":\"s2\", \
              \"resource_group\":\"rg\" \
            } \
          } \
        } \
      }"

      echo $CONFIG
  fi
else
  echo "Failed creating resource group ${SERVER_RG} for test services"
fi

exit ${RESULT}