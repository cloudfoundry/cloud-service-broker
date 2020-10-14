#!/usr/bin/env bash

set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

. "${SCRIPT_DIR}/../functions.sh"

if [ $# -lt 5 ]; then
  echo "usage: $0 <resource group> <primary server name> <secondary server name> <admin username> <admin password>"
  exit 1
fi

SERVER_RESOURCE_GROUP=$1; shift
PRIMARY_SERVER_NAME=$1; shift
SECONDARY_SERVER_NAME=$1; shift
SERVER_ADMIN_USER_NAME=$1; shift
SERVER_ADMIN_PASSWORD=$1; shift

DB_NAME=testdb-$$

RESULT=1
MASB_SQLDB_INSTANCE_NAME=mssql-db-$$
MASB_DB_CONFIG="{ \
  \"sqlServerName\": \"${PRIMARY_SERVER_NAME}\", \
  \"sqldbName\": \"${DB_NAME}\", \
  \"resourceGroup\": \"${SERVER_RESOURCE_GROUP}\" \
}"

RESULT=1
if create_service azure-sqldb StandardS0 "${MASB_SQLDB_INSTANCE_NAME}" "${MASB_DB_CONFIG}"; then
    MASB_FOG_INSTANCE_NAME=masb-fog-db-$$

    MASB_FOG_CONFIG="{ \
      \"primaryServerName\": \"${PRIMARY_SERVER_NAME}\", \
      \"primaryDbName\": \"${DB_NAME}\", \
      \"secondaryServerName\": \"${SECONDARY_SERVER_NAME}\", \
      \"failoverGroupName\": \"${MASB_FOG_INSTANCE_NAME}\", \
      \"readWriteEndpoint\": { \
        \"failoverPolicy\": \"Automatic\", \
        \"failoverWithDataLossGracePeriodMinutes\": 60 \
      } \
    }"
    if create_service azure-sqldb-failover-group SecondaryDatabaseWithFailoverGroup "${MASB_FOG_INSTANCE_NAME}" "${MASB_FOG_CONFIG}"; then
        SUBSUME_CONFIG="{ \
            \"azure_primary_db_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${PRIMARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query databases[0] -o tsv)\", \
            \"azure_secondary_db_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${SECONDARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query databases[0] -o tsv)\", \
            \"auzure_fog_id\": \"$(az sql failover-group show --name ${MASB_FOG_INSTANCE_NAME} --server ${PRIMARY_SERVER_NAME} --resource-group ${SERVER_RESOURCE_GROUP} --query id -o tsv)\", \

            \"server_pair\": \"test_server\", \
            \"server_credential_pairs\": { \
              \"test_server\": { \
                \"admin_username\":\"${SERVER_ADMIN_USER_NAME}\", \
                \"admin_password\":\"${SERVER_ADMIN_PASSWORD}\", \
                \"primary\":{ \
                    \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                    \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                }, \
                \"secondary\":{ \
                    \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                    \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                } \
              } \
            } \
          }"

        echo $SUBSUME_CONFIG

        exit 1

        SUBSUMED_INSTANCE_NAME=masb-sql-db-subsume-test-$$
        if create_service csb-masb-mssql-fog-db-subsume current "${SUBSUMED_INSTANCE_NAME}" "${SUBSUME_CONFIG}"; then
            if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}"; then
                echo "subsumed masb fog instance test successful"

                UPDATE_CONFIG="{ \
                    \"edition\": \"Standard\", \
                    \"service_objective\": \"S1\", \
                    \"server_pair\": \"test_server\", \
                    \"server_credential_pairs\": { \
                        \"test_server\": { \
                            \"admin_username\":\"${SERVER_ADMIN_USER_NAME}\", \
                            \"admin_password\":\"${SERVER_ADMIN_PASSWORD}\", \
                            \"primary\":{ \
                                \"server_name\":\"${PRIMARY_SERVER_NAME}\", \
                                \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                            }, \
                            \"secondary\":{ \
                                \"server_name\":\"${SECONDARY_SERVER_NAME}\", \
                                \"resource_group\":\"${SERVER_RESOURCE_GROUP}\" \
                            } \
                        } \
                    } \
                }"

                if update_service_params "${SUBSUMED_INSTANCE_NAME}" "${UPDATE_CONFIG}"; then
                    echo "subsumed masb fog instance update successful"      
                    if "${SCRIPT_DIR}/../cf-run-spring-music-test.sh" "${SUBSUMED_INSTANCE_NAME}"; then
                        echo "subsumed masb fog instance update test successful"           
                        RESULT=0
                    else
                        echo "updated subsumed masb fog instance test failed"
                    fi
                else
                    echo "failed to update subsumed masb fog instance"
                fi
                else
                echo "subsumed masb fog instance test failed"
            fi
            delete_service "${SUBSUMED_INSTANCE_NAME}"
            cf purge-service-instance -f "${MASB_FOG_INSTANCE_NAME}"
            cf purge-service-instance -f "${MASB_SQLDB_INSTANCE_NAME}"
        else            
            delete_service "${MASB_FOG_INSTANCE_NAME}"
        fi
    else
        delete_service "${MASB_SQLDB_INSTANCE_NAME}"
    fi
fi

exit ${RESULT}