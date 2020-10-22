#!/usr/bin/env bash

set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

echo $#

if [ $# -lt 4 ]; then
  echo "usage: $0 <server name> <resource group> <admin username> <admin password>"
  exit 1
fi

SERVER_NAME=$1
SERVER_RESOURCE_GROUP=$2
SERVER_ADMIN_USER_NAME=$3
SERVER_ADMIN_PASSWORD=$4
DB_NAME=subsume-test-db-$$
SUBSUMED_INSTANCE_NAME=masb-sql-db-subsume-test-$$

MASB_SQLDB_INSTANCE_NAME=mssql-db-$$
MASB_DB_CONFIG="{ \
  \"sqlServerName\": \"${SERVER_NAME}\", \
  \"sqldbName\": \"${DB_NAME}\", \
  \"resourceGroup\": \"${SERVER_RESOURCE_GROUP}\" \
}"

RESULT=1
if create_service azure-sqldb StandardS0 "${MASB_SQLDB_INSTANCE_NAME}" "${MASB_DB_CONFIG}"; then

  SUBSUME_CONFIG="{ \
        \"azure_db_id\": \"$(az sql db show  --name ${DB_NAME} --server ${SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query id -o tsv)\", \
        \"server\": \"test_server\", \
        \"server_credentials\": { \
          \"test_server\": { \
            \"server_name\":\"${SERVER_NAME}\", \
            \"admin_username\":\"${SERVER_ADMIN_USER_NAME}\", \
            \"admin_password\":\"${SERVER_ADMIN_PASSWORD}\", \
            \"server_resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
          }, \
          \"fail_server\": { \
            \"server_name\":\"missing\", \
            \"admin_username\":\"bogus\", \
            \"admin_password\":\"bad-password\", \
            \"server_resource_group\":\"rg\" \
          } \
        } \
      }"

  echo $SUBSUME_CONFIG

  if create_service csb-azure-mssql-db subsume "${SUBSUMED_INSTANCE_NAME}" "${SUBSUME_CONFIG}"; then
    if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}"; then
      echo "subsumed masb sqldb instance test successful"

      UPDATE_CONFIG="{ \
        \"server\": \"test_server\", \
        \"server_credentials\": { \
          \"test_server\": { \
            \"server_name\":\"${SERVER_NAME}\", \
            \"admin_username\":\"${SERVER_ADMIN_USER_NAME}\", \
            \"admin_password\":\"${SERVER_ADMIN_PASSWORD}\", \
            \"server_resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
          }, \
          \"fail_server\": { \
            \"server_name\":\"missing\", \
            \"admin_username\":\"bogus\", \
            \"admin_password\":\"bad-password\", \
            \"server_resource_group\":\"rg\" \
          } \
        } \
      }"

      if update_service_plan "${SUBSUMED_INSTANCE_NAME}" subsume "${UPDATE_CONFIG}"; then
        echo "should not have been able to update to subsume plan"
      else
        if update_service_plan "${SUBSUMED_INSTANCE_NAME}" large "${UPDATE_CONFIG}"; then
          echo "subsumed masb sqldb instance update successful"      
          if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}"; then
            echo "subsumed masb sqldb instance update test successful"           
            RESULT=0
          else
            echo "updated subsumed masb sqldb instance test failed"
          fi
        else
          echo "failed to update subsumed masb sqldb instance"
        fi
      fi
    else
      echo "subsumed masb sqldb instance test failed"
    fi
    delete_service "${SUBSUMED_INSTANCE_NAME}"
  fi

  delete_service "${MASB_SQLDB_INSTANCE_NAME}"
fi

exit ${RESULT}